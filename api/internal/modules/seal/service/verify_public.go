package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/seal/domain"
)

type PublicSeals interface {
	GetPublicSeal(ctx context.Context, code string) (domain.PublicSeal, error)
}

type VerifyPublicSeal struct {
	seals      PublicSeals
	signatures SignatureChecker
}

func NewVerifyPublicSeal(seals PublicSeals, signatures SignatureChecker) *VerifyPublicSeal {
	return &VerifyPublicSeal{seals: seals, signatures: signatures}
}

func (uc *VerifyPublicSeal) Execute(ctx context.Context, rawCode string) (domain.PublicVerdict, error) {
	code, ok := domain.NormalizeCode(rawCode)
	if !ok {
		return domain.PublicVerdict{}, apperror.NotFound()
	}
	found, err := uc.seals.GetPublicSeal(ctx, code)
	if err != nil {
		return domain.PublicVerdict{}, fmt.Errorf("get seal %s: %w", code, err)
	}
	return domain.JudgePublic(found, isSignedWith(uc.signatures)), nil
}

func isSignedWith(signatures SignatureChecker) func(domain.Seal) bool {
	return func(seal domain.Seal) bool {
		isValid, err := signatures.Verify(seal.EntryHash, seal.Signature)
		return err == nil && isValid
	}
}
