package service_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/modules/seal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/seal/service"
)

type chainTrail struct{ chain *memoryChain }

func (c chainTrail) ListEntitySeals(context.Context, domain.Document) ([]domain.TrailEntry, error) {
	trail := make([]domain.TrailEntry, 0, len(c.chain.seals))
	for _, seal := range c.chain.seals {
		trail = append(trail, domain.TrailEntry{Seal: seal})
	}
	return trail, nil
}

type stampChecker struct{}

func (stampChecker) Verify(digest []byte, signature domain.Signature) (bool, error) {
	if signature.KeyID != "v1" {
		return false, errors.New("unknown key")
	}
	return bytes.Equal(signature.Bytes, append([]byte("signed:"), digest...)), nil
}

func TestCheck_RSEC33_RecomputesTheContentHash(t *testing.T) {
	chain := &memoryChain{}
	recorder := service.NewRecorder(chain, stampSigner{}, &sequenceIDs{})
	if _, err := recorder.Seal(t.Context(), request("KS-1")); err != nil {
		t.Fatalf("seal: %v", err)
	}
	verifier := service.NewVerifier(chainTrail{chain: chain}, stampChecker{})

	same, sameErr := verifier.Check(t.Context(), domain.Check{Content: map[string]any{"version": 1}})
	changed, changedErr := verifier.Check(t.Context(), domain.Check{Content: map[string]any{"version": 2}})

	if errors.Join(sameErr, changedErr) != nil || !same.IsValid || changed.IsValid {
		t.Fatalf("same = %+v, changed = %+v", same, changed)
	}
}

func TestCheck_RSEC33_UnknownKeyIsInvalid(t *testing.T) {
	chain := &memoryChain{}
	sealed, _ := service.NewRecorder(chain, stampSigner{}, &sequenceIDs{}).Seal(t.Context(), request("KS-1"))
	chain.seals[0].Signature.KeyID = "v9"
	verifier := service.NewVerifier(chainTrail{chain: chain}, stampChecker{})

	verdict, err := verifier.Check(t.Context(), domain.Check{Content: map[string]any{"version": 1}})

	if err != nil || verdict.IsValid || sealed.Code == "" {
		t.Fatalf("verdict = %+v, err = %v", verdict, err)
	}
}
