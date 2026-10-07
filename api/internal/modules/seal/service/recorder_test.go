package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/modules/seal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/seal/service"
)

type memoryChain struct {
	seals      []domain.Seal
	failAppend error
}

func (m *memoryChain) LockHead(context.Context, string) (domain.Head, error) {
	if len(m.seals) == 0 {
		return domain.Head{}, nil
	}
	return m.seals[len(m.seals)-1].Head(), nil
}

func (m *memoryChain) Append(_ context.Context, seal domain.Seal) error {
	if m.failAppend != nil {
		return m.failAppend
	}
	m.seals = append(m.seals, seal)
	return nil
}

type stampSigner struct{}

func (stampSigner) Sign(digest []byte) domain.Signature {
	return domain.Signature{KeyID: "v1", Bytes: append([]byte("signed:"), digest...)}
}

type sequenceIDs struct{ next int }

func (s *sequenceIDs) NewID() (string, error) {
	s.next++
	return fmt.Sprintf("0192f0c4-0000-7000-8000-%012d", s.next), nil
}

func request(number string) domain.Request {
	return domain.Request{
		PartnershipID: "0192f0c4-0000-7000-8000-000000000501", Event: domain.AgreementAccepted,
		Document: domain.Document{Type: "AGREEMENT", ID: "0192f0c4-0000-7000-8000-000000000401", Number: number},
		Content:  map[string]any{"version": 1}, ActorUserID: "u", ActorOrgID: "o",
		At: time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC),
	}
}

func TestSeal_BRSIG02_EachSealChainsToTheLastAndIsSigned(t *testing.T) {
	chain := &memoryChain{}
	recorder := service.NewRecorder(chain, stampSigner{}, &sequenceIDs{})

	first, firstErr := recorder.Seal(t.Context(), request("KS-1"))
	second, secondErr := recorder.Seal(t.Context(), request("KS-2"))

	if errors.Join(firstErr, secondErr) != nil || second.Entry.Seq != 2 ||
		string(second.Entry.PrevHash) != string(first.EntryHash) || second.Signature.KeyID != "v1" ||
		domain.VerifyChain(chain.seals) != nil {
		t.Fatalf("first = %+v, second = %+v", first, second)
	}
}

func TestSeal_BRSIG03_FailedAppendFailsTheEvent(t *testing.T) {
	broken := errors.New("disk full")
	recorder := service.NewRecorder(&memoryChain{failAppend: broken}, stampSigner{}, &sequenceIDs{})

	_, err := recorder.Seal(t.Context(), request("KS-1"))

	if !errors.Is(err, broken) {
		t.Fatalf("err = %v", err)
	}
}
