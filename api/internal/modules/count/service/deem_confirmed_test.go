package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/service"
)

type unanswered struct {
	deadlines []domain.Deadline
}

func (u *unanswered) ConfirmUnanswered(_ context.Context, deadline domain.Deadline) (int, error) {
	u.deadlines = append(u.deadlines, deadline)
	return 3, nil
}

func TestDeemConfirmed_SpecTwentyOneFive_ConfirmsWhatWaitedFortyEightHours(t *testing.T) {
	store := &unanswered{}
	now := time.Date(2026, 11, 10, 17, 5, 0, 0, time.UTC)

	confirmed, err := service.NewDeemConfirmed(store).Execute(t.Context(), now)

	deadline := store.deadlines[0]
	if err != nil || confirmed != 3 || !deadline.Cutoff.Equal(now.Add(-48*time.Hour)) ||
		deadline.Window != 48*time.Hour || !deadline.At.Equal(now) {
		t.Fatalf("confirmed = %d, err = %v, deadline = %+v", confirmed, err, deadline)
	}
}
