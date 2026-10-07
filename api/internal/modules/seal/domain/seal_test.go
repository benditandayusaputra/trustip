package domain

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type vector struct {
	ContentHash    string `json:"content_hash"`
	EntryCanonical string `json:"entry_canonical"`
	EntryHash      string `json:"entry_hash"`
	Code           string `json:"code"`
	Entry          struct {
		PartnershipID string `json:"partnership_id"`
		EventType     string `json:"event_type"`
		EntityType    string `json:"entity_type"`
		EntityID      string `json:"entity_id"`
		EntityNumber  string `json:"entity_number"`
		ActorUserID   string `json:"actor_user_id"`
		ActorOrgID    string `json:"actor_org_id"`
		OccurredAt    string `json:"occurred_at"`
	} `json:"entry"`
}

func loadVector(t *testing.T) vector {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "testdata", "vectors", "seal.json"))
	if err != nil {
		t.Fatalf("read vector: %v", err)
	}
	var loaded vector
	if err := json.Unmarshal(raw, &loaded); err != nil {
		t.Fatalf("decode vector: %v", err)
	}
	return loaded
}

func vectorContent() map[string]any {
	return map[string]any{
		"agreement_id": "0192f0c4-0000-7000-8000-000000000401", "version": 1,
		"items": []any{map[string]any{
			"product_id": "0192f0c4-0000-7000-8000-000000000101", "pricing_model": "FIXED_DEPOSIT",
			"retail_price": int64(2000), "deposit_price": int64(1750),
		}},
	}
}

func vectorRequest(t *testing.T, loaded vector) Request {
	t.Helper()
	occurredAt, err := time.Parse(time.RFC3339Nano, loaded.Entry.OccurredAt)
	if err != nil {
		t.Fatalf("time: %v", err)
	}
	return Request{
		PartnershipID: loaded.Entry.PartnershipID, Event: Event(loaded.Entry.EventType),
		Document: Document{Type: loaded.Entry.EntityType, ID: loaded.Entry.EntityID, Number: loaded.Entry.EntityNumber},
		Content:  vectorContent(), ActorUserID: loaded.Entry.ActorUserID, ActorOrgID: loaded.Entry.ActorOrgID,
		At: occurredAt.Add(789 * time.Nanosecond),
	}
}

func noSignature([]byte) Signature { return Signature{KeyID: "v1"} }

func TestSealed_BRSIG01_SharedVectorGivesTheSameHashesAndCode(t *testing.T) {
	loaded := loadVector(t)
	entry, err := NewEntry(vectorRequest(t, loaded), Head{})
	if err != nil {
		t.Fatalf("entry: %v", err)
	}

	canonical, canonicalErr := entry.Canonical()
	seal, sealErr := Sealed("s1", entry, noSignature)

	if errors.Join(canonicalErr, sealErr) != nil || hex.EncodeToString(entry.ContentHash) != loaded.ContentHash ||
		string(canonical) != loaded.EntryCanonical || hex.EncodeToString(seal.EntryHash) != loaded.EntryHash ||
		seal.Code != loaded.Code {
		t.Fatalf("canonical = %s\nseal = %+v", canonical, seal)
	}
}

func TestNewEntry_BRSIG02_ChainsToThePreviousHead(t *testing.T) {
	loaded := loadVector(t)
	head := Head{Seq: 4, Hash: []byte{1, 2, 3}}

	entry, err := NewEntry(vectorRequest(t, loaded), head)

	if err != nil || entry.Seq != 5 || string(entry.PrevHash) != string(head.Hash) {
		t.Fatalf("entry = %+v, err = %v", entry, err)
	}
}

func TestCheckEntry_RSEC33_OneChangedCharacterIsDetected(t *testing.T) {
	loaded := loadVector(t)
	entry, _ := NewEntry(vectorRequest(t, loaded), Head{})
	seal, _ := Sealed("s1", entry, noSignature)

	seal.Entry.Document.Number = "KS-2"

	if !errors.Is(seal.CheckEntry(), ErrEntryMismatch) {
		t.Fatal("changed entry passed")
	}
}

func chainOf(t *testing.T, length int) []Seal {
	t.Helper()
	request := vectorRequest(t, loadVector(t))
	seals := []Seal{}
	var head Head
	for range length {
		entry, err := NewEntry(request, head)
		if err != nil {
			t.Fatalf("entry: %v", err)
		}
		seal, err := Sealed("s", entry, noSignature)
		if err != nil {
			t.Fatalf("seal: %v", err)
		}
		seals = append(seals, seal)
		head = seal.Head()
	}
	return seals
}

func TestVerifyChain_FRSIG05_IntactChainPasses(t *testing.T) {
	if err := VerifyChain(chainOf(t, 3)); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestVerifyChain_FRSIG05_RemovedEntryBreaksTheChain(t *testing.T) {
	seals := chainOf(t, 3)

	err := VerifyChain([]Seal{seals[0], seals[2]})

	if !errors.Is(err, ErrBrokenChain) {
		t.Fatalf("err = %v", err)
	}
}

func TestVerifyChain_FRSIG05_RewrittenEntryIsDetected(t *testing.T) {
	seals := chainOf(t, 2)
	seals[1].Entry.ActorOrgID = "someone-else"

	if !errors.Is(VerifyChain(seals), ErrEntryMismatch) {
		t.Fatal("rewritten entry passed")
	}
}

func TestCode_BRSIG05_SixteenBase32CharactersInGroupsOfFour(t *testing.T) {
	code := Code(make([]byte, 32))

	if code != "AAAA-AAAA-AAAA-AAAA" {
		t.Fatalf("code = %s", code)
	}
}

func TestNewEntry_RSEC28_ContentOutsideTheCanonicalTypesIsRefused(t *testing.T) {
	request := vectorRequest(t, loadVector(t))
	request.Content = map[string]any{"price": 1.5}

	if _, err := NewEntry(request, Head{}); err == nil {
		t.Fatal("fractional content accepted")
	}
}

func TestSealed_RSEC28_InvalidTextInTheEntryIsRefused(t *testing.T) {
	entry, _ := NewEntry(vectorRequest(t, loadVector(t)), Head{})
	entry.Document.Number = "\xff"

	_, sealErr := Sealed("s1", entry, noSignature)
	checkErr := Seal{Entry: entry}.CheckEntry()

	if sealErr == nil || checkErr == nil || errors.Is(checkErr, ErrEntryMismatch) {
		t.Fatalf("seal = %v, check = %v", sealErr, checkErr)
	}
}
