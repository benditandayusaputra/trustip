package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/canonical"
)

const (
	FormatVersion   = 1
	codeLength      = 16
	codeGroupLength = 4
	codeSeparator   = "-"
	instantLayout   = "2006-01-02T15:04:05.000000Z"
)

var (
	ErrBrokenChain   = errors.New("seal chain is broken")
	ErrEntryMismatch = errors.New("seal entry does not match its hash")
)

type Event string

const (
	AgreementAccepted Event = "AGREEMENT_ACCEPTED"
	DropRecorded      Event = "DROP_RECORDED"
	DropConfirmed     Event = "DROP_CONFIRMED"
	DropCancelled     Event = "DROP_CANCELLED"
	CountRecorded     Event = "COUNT_RECORDED"
	CountConfirmed    Event = "COUNT_CONFIRMED"
	SettlementIssued  Event = "SETTLEMENT_ISSUED"
	PaymentRecorded   Event = "PAYMENT_RECORDED"
	PaymentConfirmed  Event = "PAYMENT_CONFIRMED"
	PaymentRejected   Event = "PAYMENT_REJECTED"
	PartnershipEnded  Event = "PARTNERSHIP_ENDED"
)

type Document struct {
	Type   string
	ID     string
	Number string
}

type Request struct {
	PartnershipID string
	Event         Event
	Document      Document
	Content       any
	ActorUserID   string
	ActorOrgID    string
	At            time.Time
}

type Head struct {
	Seq  int
	Hash []byte
}

type Entry struct {
	PartnershipID string
	Seq           int
	Event         Event
	Document      Document
	ContentHash   []byte
	ActorUserID   string
	ActorOrgID    string
	OccurredAt    time.Time
	PrevHash      []byte
}

func NewEntry(request Request, head Head) (Entry, error) {
	contentHash, err := canonical.Digest(request.Content)
	if err != nil {
		return Entry{}, fmt.Errorf("hash %s content: %w", request.Document.Type, err)
	}
	return Entry{
		PartnershipID: request.PartnershipID, Seq: head.Seq + 1, Event: request.Event, Document: request.Document,
		ContentHash: contentHash, ActorUserID: request.ActorUserID, ActorOrgID: request.ActorOrgID,
		OccurredAt: request.At.UTC().Truncate(time.Microsecond), PrevHash: head.Hash,
	}, nil
}

func (e Entry) Canonical() ([]byte, error) {
	encoded, err := canonical.Encode(map[string]any{
		"v": FormatVersion, "partnership_id": e.PartnershipID, "seq": e.Seq, "event_type": string(e.Event),
		"entity_type": e.Document.Type, "entity_id": e.Document.ID, "entity_number": e.Document.Number,
		"content_hash": hex.EncodeToString(e.ContentHash), "actor_user_id": e.ActorUserID,
		"actor_org_id": e.ActorOrgID, "occurred_at": e.OccurredAt.UTC().Format(instantLayout),
		"prev_hash": hex.EncodeToString(e.PrevHash),
	})
	if err != nil {
		return nil, fmt.Errorf("encode seal entry: %w", err)
	}
	return encoded, nil
}

func (e Entry) Hash() ([]byte, error) {
	encoded, err := e.Canonical()
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(encoded)
	return sum[:], nil
}

func Code(entryHash []byte) string {
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(entryHash)[:codeLength]
	groups := make([]string, 0, codeLength/codeGroupLength)
	for start := 0; start < codeLength; start += codeGroupLength {
		groups = append(groups, encoded[start:start+codeGroupLength])
	}
	return strings.Join(groups, codeSeparator)
}

type Signature struct {
	KeyID string
	Bytes []byte
}

type Seal struct {
	ID        string
	Entry     Entry
	EntryHash []byte
	Signature Signature
	Code      string
}

func Sealed(id string, entry Entry, sign func(digest []byte) Signature) (Seal, error) {
	entryHash, err := entry.Hash()
	if err != nil {
		return Seal{}, err
	}
	return Seal{ID: id, Entry: entry, EntryHash: entryHash, Signature: sign(entryHash), Code: Code(entryHash)}, nil
}

func (s Seal) Head() Head {
	return Head{Seq: s.Entry.Seq, Hash: s.EntryHash}
}

func (s Seal) CheckEntry() error {
	recomputed, err := s.Entry.Hash()
	if err != nil {
		return err
	}
	if !bytes.Equal(recomputed, s.EntryHash) {
		return ErrEntryMismatch
	}
	return nil
}

func VerifyChain(seals []Seal) error {
	var previous Head
	for _, seal := range seals {
		if err := seal.CheckEntry(); err != nil {
			return fmt.Errorf("seal %d: %w", seal.Entry.Seq, err)
		}
		if seal.Entry.Seq != previous.Seq+1 || !bytes.Equal(seal.Entry.PrevHash, previous.Hash) {
			return fmt.Errorf("seal %d: %w", seal.Entry.Seq, ErrBrokenChain)
		}
		previous = seal.Head()
	}
	return nil
}

type TrailEntry struct {
	Seal          Seal
	ActorOrgName  string
	ActorUserName string
}

type Check struct {
	Document Document
	Content  any
}

type Verdict struct {
	IsValid bool
	Trail   []TrailEntry
}

func Judge(trail []TrailEntry, contentHash []byte, isSigned func(Seal) bool) Verdict {
	if len(trail) == 0 {
		return Verdict{Trail: trail}
	}
	latest := trail[len(trail)-1].Seal
	isValid := bytes.Equal(latest.Entry.ContentHash, contentHash)
	for _, entry := range trail {
		isValid = isValid && entry.Seal.CheckEntry() == nil && isSigned(entry.Seal)
	}
	return Verdict{IsValid: isValid, Trail: trail}
}
