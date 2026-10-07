package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	"github.com/benditandayusaputra/trustip/api/internal/modules/agreement/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/agreement/service"
)

type memoryAgreements struct {
	partnerships []domain.Partnership
	agreements   []domain.Agreement
	failPartner  error
}

func (m *memoryAgreements) CreatePartnership(_ context.Context, partnership domain.Partnership) error {
	if m.failPartner != nil {
		return m.failPartner
	}
	m.partnerships = append(m.partnerships, partnership)
	return nil
}

func (m *memoryAgreements) CreateAgreement(_ context.Context, agreement domain.Agreement) error {
	m.agreements = append(m.agreements, agreement)
	return nil
}

type memorySealer struct{ requests []domain.SealRequest }

func (m *memorySealer) Seal(_ context.Context, request domain.SealRequest) error {
	m.requests = append(m.requests, request)
	return nil
}

type sequenceIDs struct{ next int }

func (s *sequenceIDs) NewID() (string, error) {
	s.next++
	return fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", s.next), nil
}

func input() domain.FoundingInput {
	return domain.FoundingInput{
		ProposalID: "p1", ProposalNumber: "AJ-202610-0001", ConsignorID: "c1", StoreID: "s1",
		Terms:      terms.General{Schedule: terms.ScheduleOnCount, Rhythm: terms.RhythmDaily},
		ProposedBy: "u1", AcceptedBy: "u2", AcceptorOrgID: "s1", At: time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC),
	}
}

func TestFound_BRPRP03_CreatesPartnershipAgreementAndSealTogether(t *testing.T) {
	agreements, sealer := &memoryAgreements{}, &memorySealer{}
	founder := service.NewFounder(service.FounderDeps{Agreements: agreements, Sealer: sealer, IDs: &sequenceIDs{}})

	agreement, err := founder.Found(t.Context(), input())

	if err != nil || len(agreements.partnerships) != 1 || len(agreements.agreements) != 1 ||
		len(sealer.requests) != 1 ||
		sealer.requests[0].Event != "AGREEMENT_ACCEPTED" || sealer.requests[0].ActorOrgID != "s1" ||
		sealer.requests[0].DocumentID != agreement.ID {
		t.Fatalf("agreement = %+v, sealer = %+v, err = %v", agreement, sealer.requests, err)
	}
}

func TestFound_BRPRT01_ExistingActivePartnershipStopsTheFounding(t *testing.T) {
	agreements, sealer := &memoryAgreements{failPartner: domain.ErrAlreadyPartners}, &memorySealer{}
	founder := service.NewFounder(service.FounderDeps{Agreements: agreements, Sealer: sealer, IDs: &sequenceIDs{}})

	_, err := founder.Found(t.Context(), input())

	if !errors.Is(err, domain.ErrAlreadyPartners) || len(sealer.requests) != 0 {
		t.Fatalf("err = %v", err)
	}
}
