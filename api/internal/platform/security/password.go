package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const (
	argon2SaltBytes        = 16
	argon2KeyBytes         = 32
	maxConcurrentHashes    = 4
	busyRetryAfterSeconds  = 1
	minimumArgon2MemoryKiB = 19 * 1024
	minimumIterations      = 2
	minimumParallelism     = 1
)

var (
	ErrWeakArgon2Params      = errors.New("argon2 parameters are below the required minimum")
	ErrUnknownPepper         = errors.New("password pepper version is unknown")
	ErrMalformedPasswordHash = errors.New("stored password hash is malformed")
)

type Argon2Params struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
}

func MinimumArgon2Params() Argon2Params {
	return Argon2Params{
		MemoryKiB:   minimumArgon2MemoryKiB,
		Iterations:  minimumIterations,
		Parallelism: minimumParallelism,
	}
}

func (p Argon2Params) isWeakerThan(other Argon2Params) bool {
	return p.MemoryKiB < other.MemoryKiB || p.Iterations < other.Iterations || p.Parallelism < other.Parallelism
}

type PepperRing interface {
	Active() (string, []byte)
	Lookup(id string) ([]byte, bool)
}

type StoredPassword struct {
	Hash          string
	PepperVersion string
}

type PasswordCheck struct {
	Matches     bool
	NeedsRehash bool
}

type PasswordHasher struct {
	peppers PepperRing
	params  Argon2Params
	slots   chan struct{}
}

func NewPasswordHasher(peppers PepperRing, params Argon2Params) (*PasswordHasher, error) {
	if params.isWeakerThan(MinimumArgon2Params()) {
		return nil, ErrWeakArgon2Params
	}
	if _, pepper := peppers.Active(); len(pepper) == 0 {
		return nil, ErrUnknownPepper
	}
	return &PasswordHasher{peppers: peppers, params: params, slots: make(chan struct{}, maxConcurrentHashes)}, nil
}

func (h *PasswordHasher) Hash(password string) (StoredPassword, error) {
	salt := make([]byte, argon2SaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return StoredPassword{}, fmt.Errorf("read password salt: %w", err)
	}
	pepperVersion, pepper := h.peppers.Active()
	phc := phcHash{params: h.params, salt: salt}
	key, err := h.derive(phc, pepper, password)
	if err != nil {
		return StoredPassword{}, err
	}
	phc.key = key
	return StoredPassword{Hash: phc.String(), PepperVersion: pepperVersion}, nil
}

func (h *PasswordHasher) Verify(stored StoredPassword, password string) (PasswordCheck, error) {
	phc, err := parsePHC(stored.Hash)
	if err != nil {
		return PasswordCheck{}, err
	}
	pepper, ok := h.peppers.Lookup(stored.PepperVersion)
	if !ok {
		return PasswordCheck{}, ErrUnknownPepper
	}
	key, err := h.derive(phc, pepper, password)
	if err != nil {
		return PasswordCheck{}, err
	}
	matches := subtle.ConstantTimeCompare(key, phc.key) == 1
	return PasswordCheck{Matches: matches, NeedsRehash: matches && h.isOutdated(stored, phc)}, nil
}

func (h *PasswordHasher) VerifyAbsent(password string) error {
	_, pepper := h.peppers.Active()
	absent := phcHash{params: h.params, salt: make([]byte, argon2SaltBytes)}
	_, err := h.derive(absent, pepper, password)
	return err
}

func (h *PasswordHasher) isOutdated(stored StoredPassword, phc phcHash) bool {
	activeVersion, _ := h.peppers.Active()
	return stored.PepperVersion != activeVersion || phc.params.isWeakerThan(h.params)
}

func (h *PasswordHasher) derive(phc phcHash, pepper []byte, password string) ([]byte, error) {
	select {
	case h.slots <- struct{}{}:
	default:
		return nil, apperror.Unavailable(busyRetryAfterSeconds)
	}
	defer func() { <-h.slots }()
	mac := hmac.New(sha256.New, pepper)
	mac.Write([]byte(password))
	return argon2.IDKey(mac.Sum(nil), phc.salt, phc.params.Iterations, phc.params.MemoryKiB,
		phc.params.Parallelism, argon2KeyBytes), nil
}
