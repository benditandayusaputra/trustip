package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/canonical"
	"github.com/benditandayusaputra/trustip/api/internal/modules/seal/domain"
)

type TrailReader interface {
	ListEntitySeals(ctx context.Context, document domain.Document) ([]domain.TrailEntry, error)
}

type SignatureChecker interface {
	Verify(digest []byte, signature domain.Signature) (bool, error)
}

type Verifier struct {
	trails     TrailReader
	signatures SignatureChecker
}

func NewVerifier(trails TrailReader, signatures SignatureChecker) *Verifier {
	return &Verifier{trails: trails, signatures: signatures}
}

func (v *Verifier) Check(ctx context.Context, check domain.Check) (domain.Verdict, error) {
	trail, err := v.trails.ListEntitySeals(ctx, check.Document)
	if err != nil {
		return domain.Verdict{}, fmt.Errorf("list seals of %s %s: %w", check.Document.Type, check.Document.ID, err)
	}
	contentHash, err := canonical.Digest(check.Content)
	if err != nil {
		return domain.Verdict{}, fmt.Errorf("hash %s content: %w", check.Document.Type, err)
	}
	return domain.Judge(trail, contentHash, v.isSigned), nil
}

func (v *Verifier) isSigned(seal domain.Seal) bool {
	isValid, err := v.signatures.Verify(seal.EntryHash, seal.Signature)
	return err == nil && isValid
}
