package service

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
)

type Issuer struct {
	deps Deps
}

func NewIssuer(deps Deps) *Issuer {
	return &Issuer{deps: deps}
}

func (i *Issuer) IssueOnCount(
	ctx context.Context, partnershipID string, countedAt time.Time,
) (domain.Settlement, bool, error) {
	ledger, err := i.deps.Settlements.GetLedger(ctx, partnershipID)
	if err != nil {
		return domain.Settlement{}, false, fmt.Errorf("load ledger of %s: %w", partnershipID, err)
	}
	if ledger.Schedule != terms.ScheduleOnCount {
		return domain.Settlement{}, false, nil
	}
	draft, hasSales := ledger.DraftUpTo(countedAt)
	if !hasSales {
		return domain.Settlement{}, false, nil
	}
	settlement, err := i.issue(ctx, ledger, draft)
	if err != nil {
		return domain.Settlement{}, false, err
	}
	return settlement, true, nil
}

func (i *Issuer) issue(ctx context.Context, ledger domain.Ledger, draft domain.Draft) (domain.Settlement, error) {
	id, err := i.deps.IDs.NewID()
	if err != nil {
		return domain.Settlement{}, fmt.Errorf("generate settlement id: %w", err)
	}
	now := i.deps.Clock.Now()
	sequence, err := i.deps.Settlements.NextNumber(ctx, domain.SequenceOf(ledger.ConsignorID, now))
	if err != nil {
		return domain.Settlement{}, fmt.Errorf("number settlement: %w", err)
	}
	settlement := draft.Issue(domain.Issuing{ID: id, Sequence: sequence, At: now})
	if err := i.deps.Settlements.CreateSettlement(ctx, settlement); err != nil {
		return domain.Settlement{}, fmt.Errorf("save settlement %s: %w", settlement.Number, err)
	}
	if err := i.deps.Sealer.Seal(ctx, settlement.SealRequest()); err != nil {
		return domain.Settlement{}, fmt.Errorf("seal settlement %s: %w", settlement.Number, err)
	}
	if err := i.tell(ctx, settlement, ledger.StoreName); err != nil {
		return domain.Settlement{}, err
	}
	return settlement, nil
}

func (i *Issuer) tell(ctx context.Context, settlement domain.Settlement, storeName string) error {
	for _, notice := range settlement.Notices(storeName) {
		if err := i.deps.Announcer.Announce(ctx, notice); err != nil {
			return fmt.Errorf("announce settlement %s: %w", settlement.Number, err)
		}
		err := i.deps.Audit.Record(ctx, audit.Entry{
			OrganizationID: notice.OrganizationID, Action: audit.SettlementIssued,
			EntityType: audit.SettlementEntity, EntityID: settlement.ID, At: settlement.IssuedAt,
		})
		if err != nil {
			return fmt.Errorf("audit settlement %s: %w", settlement.Number, err)
		}
	}
	return nil
}
