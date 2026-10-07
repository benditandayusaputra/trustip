package security_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

type pepperRing struct {
	activeID string
	peppers  map[string][]byte
}

func (r pepperRing) Active() (string, []byte) {
	return r.activeID, r.peppers[r.activeID]
}

func (r pepperRing) Lookup(id string) ([]byte, bool) {
	pepper, ok := r.peppers[id]
	return pepper, ok
}

func pepper(fill byte) []byte {
	return bytes.Repeat([]byte{fill}, 32)
}

func onePepper() pepperRing {
	return pepperRing{activeID: "v1", peppers: map[string][]byte{"v1": pepper(1)}}
}

func newHasher(t *testing.T, ring pepperRing, params security.Argon2Params) *security.PasswordHasher {
	t.Helper()
	hasher, err := security.NewPasswordHasher(ring, params)
	if err != nil {
		t.Fatalf("new hasher: %v", err)
	}
	return hasher
}

const typedInput = "Kp7#mQz2vX"

func hashOf(t *testing.T, hasher *security.PasswordHasher) security.StoredPassword {
	t.Helper()
	stored, err := hasher.Hash(typedInput)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	return stored
}

func verify(t *testing.T, hasher *security.PasswordHasher, stored security.StoredPassword) security.PasswordCheck {
	t.Helper()
	check, err := hasher.Verify(stored, typedInput)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	return check
}

func TestHash_RSEC04_StoresPHCStringWithParamsAndPepperVersion(t *testing.T) {
	stored := hashOf(t, newHasher(t, onePepper(), security.MinimumArgon2Params()))

	if !strings.HasPrefix(stored.Hash, "$argon2id$v=19$m=19456,t=2,p=1$") || stored.PepperVersion != "v1" {
		t.Fatalf("stored = %+v", stored)
	}
}

func TestHash_BRAUTH04_SaltDiffersForTheSamePassword(t *testing.T) {
	hasher := newHasher(t, onePepper(), security.MinimumArgon2Params())

	first := hashOf(t, hasher)
	second := hashOf(t, hasher)

	if first.Hash == second.Hash {
		t.Fatal("BR-AUTH-04 two hashes of one password must use different salts")
	}
}

func TestVerify_CorrectPassword_MatchesWithoutRehash(t *testing.T) {
	hasher := newHasher(t, onePepper(), security.MinimumArgon2Params())

	check := verify(t, hasher, hashOf(t, hasher))

	if !check.Matches || check.NeedsRehash {
		t.Fatalf("check = %+v", check)
	}
}

func TestVerify_WrongPassword_DoesNotMatch(t *testing.T) {
	hasher := newHasher(t, onePepper(), security.MinimumArgon2Params())

	check, err := hasher.Verify(hashOf(t, hasher), "Kp7#mQz2vY")

	if err != nil || check.Matches {
		t.Fatal("a different password must not match")
	}
}

func TestVerify_BRAUTH04_WrongPepperDoesNotMatch(t *testing.T) {
	stored := hashOf(t, newHasher(t, onePepper(), security.MinimumArgon2Params()))
	otherPepper := pepperRing{activeID: "v1", peppers: map[string][]byte{"v1": pepper(9)}}

	check := verify(t, newHasher(t, otherPepper, security.MinimumArgon2Params()), stored)

	if check.Matches {
		t.Fatal("BR-AUTH-04 a hash made with another pepper must not verify")
	}
}

func TestVerify_UnknownPepperVersion_IsAnError(t *testing.T) {
	hasher := newHasher(t, onePepper(), security.MinimumArgon2Params())
	stored := hashOf(t, hasher)
	stored.PepperVersion = "v7"

	if _, err := hasher.Verify(stored, "Kp7#mQz2vX"); !errors.Is(err, security.ErrUnknownPepper) {
		t.Fatalf("err = %v", err)
	}
}

func TestVerify_MalformedHash_IsAnError(t *testing.T) {
	hasher := newHasher(t, onePepper(), security.MinimumArgon2Params())
	cases := []string{
		"",
		"$2a$10$bcrypt",
		"$argon2id$v=19$m=x,t=2,p=1$c2FsdA$aGFzaA",
		"$argon2id$v=19$m=19456,t=2,p=1$!$!",
		"$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA",
	}
	for _, hash := range cases {
		stored := security.StoredPassword{Hash: hash, PepperVersion: "v1"}
		if _, err := hasher.Verify(stored, "Kp7#mQz2vX"); !errors.Is(err, security.ErrMalformedPasswordHash) {
			t.Errorf("%q: err = %v", hash, err)
		}
	}
}

func TestVerify_RSEC03_OlderParamsAskForRehash(t *testing.T) {
	stored := hashOf(t, newHasher(t, onePepper(), security.MinimumArgon2Params()))
	stronger := security.MinimumArgon2Params()
	stronger.Iterations++

	check := verify(t, newHasher(t, onePepper(), stronger), stored)

	if !check.Matches || !check.NeedsRehash {
		t.Fatalf("check = %+v", check)
	}
}

func TestVerify_RSEC05_RotatedPepperVerifiesOldHashAndAsksForRehash(t *testing.T) {
	stored := hashOf(t, newHasher(t, onePepper(), security.MinimumArgon2Params()))
	rotated := pepperRing{activeID: "v2", peppers: map[string][]byte{"v1": pepper(1), "v2": pepper(2)}}
	hasher := newHasher(t, rotated, security.MinimumArgon2Params())

	check := verify(t, hasher, stored)
	rehashed := hashOf(t, hasher)

	if !check.Matches || !check.NeedsRehash || rehashed.PepperVersion != "v2" {
		t.Fatalf("check = %+v, rehashed pepper = %s", check, rehashed.PepperVersion)
	}
}

func TestVerifyAbsent_RSEC06_RunsWithoutError(t *testing.T) {
	hasher := newHasher(t, onePepper(), security.MinimumArgon2Params())

	if err := hasher.VerifyAbsent("Kp7#mQz2vX"); err != nil {
		t.Fatalf("verify absent: %v", err)
	}
}

func TestNewPasswordHasher_RSEC03_WeakerThanMinimumIsRejected(t *testing.T) {
	weak := []func(*security.Argon2Params){
		func(p *security.Argon2Params) { p.MemoryKiB-- },
		func(p *security.Argon2Params) { p.Iterations-- },
		func(p *security.Argon2Params) { p.Parallelism-- },
	}
	for _, weaken := range weak {
		params := security.MinimumArgon2Params()
		weaken(&params)
		if _, err := security.NewPasswordHasher(onePepper(), params); !errors.Is(err, security.ErrWeakArgon2Params) {
			t.Errorf("params %+v: err = %v", params, err)
		}
	}
}

func TestNewPasswordHasher_ActivePepperMissing_IsRejected(t *testing.T) {
	ring := pepperRing{activeID: "v2", peppers: map[string][]byte{"v1": pepper(1)}}

	_, err := security.NewPasswordHasher(ring, security.MinimumArgon2Params())

	if !errors.Is(err, security.ErrUnknownPepper) {
		t.Fatalf("err = %v", err)
	}
}
