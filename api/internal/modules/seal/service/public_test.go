package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/seal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/seal/service"
)

type publicShelf struct {
	chain *memoryChain
	asked []string
}

func (s *publicShelf) GetPublicSeal(_ context.Context, code string) (domain.PublicSeal, error) {
	s.asked = append(s.asked, code)
	for index, seal := range s.chain.seals {
		if seal.Code != code {
			continue
		}
		found := domain.PublicSeal{Seal: seal, Parties: domain.Parties{ConsignorName: "Roti", StoreName: "Warung"}}
		if index > 0 {
			found.PreviousEntryHash = s.chain.seals[index-1].EntryHash
		}
		return found, nil
	}
	return domain.PublicSeal{}, apperror.NotFound()
}

func (s *publicShelf) ListSealedPartnerships(context.Context) ([]string, error) {
	return []string{"partnership-1"}, nil
}

func (s *publicShelf) ListChain(context.Context, string) ([]domain.Seal, error) {
	return s.chain.seals, nil
}

func sealedShelf(t *testing.T, count int) *publicShelf {
	t.Helper()
	chain := &memoryChain{}
	recorder := service.NewRecorder(chain, stampSigner{}, &sequenceIDs{})
	for range count {
		if _, err := recorder.Seal(t.Context(), request("TM-1")); err != nil {
			t.Fatalf("seal: %v", err)
		}
	}
	return &publicShelf{chain: chain}
}

func TestVerifyPublicSeal_FRSIG03_KnownCodeInAnyCaseIsValid(t *testing.T) {
	shelf := sealedShelf(t, 2)
	code := shelf.chain.seals[1].Code

	verdict, err := service.NewVerifyPublicSeal(shelf, stampChecker{}).Execute(t.Context(), " "+strings.ToLower(code))
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	if !verdict.IsValid || verdict.Parties.StoreName != "Warung" {
		t.Fatalf("verdict = %+v", verdict)
	}
}

func TestVerifyPublicSeal_FRSIG03_UnknownCodeIsNotFound(t *testing.T) {
	_, err := service.NewVerifyPublicSeal(sealedShelf(t, 1), stampChecker{}).Execute(t.Context(),
		"AAAA-AAAA-AAAA-AAAA")

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestVerifyPublicSeal_MalformedCodeIsNotFoundWithoutLookup(t *testing.T) {
	shelf := sealedShelf(t, 1)

	_, err := service.NewVerifyPublicSeal(shelf, stampChecker{}).Execute(t.Context(), "bukan-kode")

	if apperror.From(err).Code() != apperror.CodeNotFound || len(shelf.asked) != 0 {
		t.Fatalf("err = %v, asked = %v", err, shelf.asked)
	}
}

func TestVerifyPublicSeal_RSEC33_TamperedEntryIsInvalid(t *testing.T) {
	shelf := sealedShelf(t, 1)
	shelf.chain.seals[0].Entry.Document.Number = "TM-9"

	verdict, err := service.NewVerifyPublicSeal(shelf, stampChecker{}).Execute(t.Context(), shelf.chain.seals[0].Code)

	if err != nil || verdict.IsValid {
		t.Fatalf("verdict = %+v, err = %v", verdict, err)
	}
}

func TestCheckChains_FRSIG05_ReportsTheBrokenPartnership(t *testing.T) {
	intact := sealedShelf(t, 3)
	broken := sealedShelf(t, 3)
	broken.chain.seals = append(broken.chain.seals[:1], broken.chain.seals[2:]...)

	good, goodErr := service.NewCheckChains(intact, stampChecker{}).Execute(t.Context())
	bad, badErr := service.NewCheckChains(broken, stampChecker{}).Execute(t.Context())

	if goodErr != nil || badErr != nil || good.Checked != 1 || len(good.Broken) != 0 ||
		len(bad.Broken) != 1 || bad.Broken[0] != "partnership-1" {
		t.Fatalf("good = %+v, bad = %+v", good, bad)
	}
}
