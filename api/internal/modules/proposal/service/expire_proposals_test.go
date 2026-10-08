package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/proposal/service"
)

type expiring struct {
	at []time.Time
}

func (e *expiring) ExpireOpen(_ context.Context, now time.Time) (int, error) {
	e.at = append(e.at, now)
	return 2, nil
}

func TestExpireProposals_SpecTwentyOneFive_MarksOpenProposalsPastTheirDeadline(t *testing.T) {
	store := &expiring{}
	now := time.Date(2026, 11, 10, 17, 5, 0, 0, time.UTC)

	expired, err := service.NewExpireProposals(store).Execute(t.Context(), now)

	if err != nil || expired != 2 || !store.at[0].Equal(now) {
		t.Fatalf("expired = %d, err = %v", expired, err)
	}
}
