package security

import (
	"crypto/ed25519"
	"errors"
)

var ErrSealUnknownKey = errors.New("seal key id is unknown")

type ActiveKeyring interface {
	KeyLookup
	Active() (string, []byte)
	IDs() []string
}

type PublicKey struct {
	KeyID string
	Bytes []byte
}

type Signature struct {
	KeyID string
	Bytes []byte
}

type Sealer struct {
	keys ActiveKeyring
}

func NewSealer(keys ActiveKeyring) *Sealer {
	return &Sealer{keys: keys}
}

func (s *Sealer) Sign(digest []byte) Signature {
	keyID, seed := s.keys.Active()
	return Signature{KeyID: keyID, Bytes: ed25519.Sign(ed25519.NewKeyFromSeed(seed), digest)}
}

func (s *Sealer) Verify(digest []byte, signature Signature) (bool, error) {
	seed, ok := s.keys.Lookup(signature.KeyID)
	if !ok {
		return false, ErrSealUnknownKey
	}
	return ed25519.Verify(publicKeyOf(seed), digest, signature.Bytes), nil
}

func (s *Sealer) PublicKey(keyID string) ([]byte, bool) {
	seed, ok := s.keys.Lookup(keyID)
	if !ok {
		return nil, false
	}
	return publicKeyOf(seed), true
}

func (s *Sealer) PublicKeys() []PublicKey {
	ids := s.keys.IDs()
	keys := make([]PublicKey, 0, len(ids))
	for _, keyID := range ids {
		if public, ok := s.PublicKey(keyID); ok {
			keys = append(keys, PublicKey{KeyID: keyID, Bytes: public})
		}
	}
	return keys
}

func publicKeyOf(seed []byte) ed25519.PublicKey {
	public, _ := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
	return public
}
