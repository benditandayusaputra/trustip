package service

import (
	"context"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/modules/seal/domain"
)

type Chains interface {
	ListSealedPartnerships(ctx context.Context) ([]string, error)
	ListChain(ctx context.Context, partnershipID string) ([]domain.Seal, error)
}

type ChainReport struct {
	Checked int
	Broken  []string
}

type CheckChains struct {
	chains     Chains
	signatures SignatureChecker
}

func NewCheckChains(chains Chains, signatures SignatureChecker) *CheckChains {
	return &CheckChains{chains: chains, signatures: signatures}
}

func (uc *CheckChains) Execute(ctx context.Context) (ChainReport, error) {
	partnerships, err := uc.chains.ListSealedPartnerships(ctx)
	if err != nil {
		return ChainReport{}, fmt.Errorf("list sealed partnerships: %w", err)
	}
	report := ChainReport{Broken: []string{}}
	isSigned := isSignedWith(uc.signatures)
	for _, partnershipID := range partnerships {
		chain, err := uc.chains.ListChain(ctx, partnershipID)
		if err != nil {
			return report, fmt.Errorf("list seal chain of %s: %w", partnershipID, err)
		}
		report.Checked++
		if !domain.IsIntactChain(chain, isSigned) {
			report.Broken = append(report.Broken, partnershipID)
		}
	}
	return report, nil
}
