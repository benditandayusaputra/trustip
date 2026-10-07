package security

import (
	"bytes"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

type singlePepper struct{}

func (singlePepper) Active() (string, []byte) {
	return "v1", bytes.Repeat([]byte{1}, 32)
}

func (singlePepper) Lookup(string) ([]byte, bool) {
	return bytes.Repeat([]byte{1}, 32), true
}

func TestHash_RSEC07_FifthConcurrentHashIsUnavailable(t *testing.T) {
	hasher, err := NewPasswordHasher(singlePepper{}, MinimumArgon2Params())
	if err != nil {
		t.Fatalf("new hasher: %v", err)
	}
	for range maxConcurrentHashes {
		hasher.slots <- struct{}{}
	}

	_, hashErr := hasher.Hash("Kp7#mQz2vX")
	absentErr := hasher.VerifyAbsent("Kp7#mQz2vX")

	for _, got := range []error{hashErr, absentErr} {
		appErr := apperror.From(got)
		if appErr.Code() != apperror.CodeServiceUnavailable || appErr.RetryAfter() < 1 {
			t.Fatalf("err = %v, retry after = %d", got, appErr.RetryAfter())
		}
	}
}

func TestHash_RSEC07_SlotIsReleasedAfterHashing(t *testing.T) {
	hasher, err := NewPasswordHasher(singlePepper{}, MinimumArgon2Params())
	if err != nil {
		t.Fatalf("new hasher: %v", err)
	}

	for range maxConcurrentHashes + 1 {
		if _, err := hasher.Hash("Kp7#mQz2vX"); err != nil {
			t.Fatalf("hash: %v", err)
		}
	}

	if len(hasher.slots) != 0 {
		t.Fatalf("slots held = %d", len(hasher.slots))
	}
}
