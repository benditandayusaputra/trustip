package domain

import (
	"fmt"
	"strconv"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/canonical"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
)

const (
	MinItems        = 1
	MaxItems        = 50
	MinQty          = 1
	MaxQty          = 10000
	maxDaysBack     = 3
	futureTolerance = 5 * time.Minute
)

type ItemInput struct {
	ProductID  string
	Qty        int
	ExpiryDate string
}

type RecordInput struct {
	PartnershipID    string
	AgreementVersion int
	OccurredAt       time.Time
	Items            []ItemInput
}

func (i RecordInput) Fingerprint() ([]byte, error) {
	items := make([]any, 0, len(i.Items))
	for _, item := range i.Items {
		items = append(items, map[string]any{
			"product_id": item.ProductID, "qty": item.Qty, "expiry_date": item.ExpiryDate,
		})
	}
	digest, err := canonical.Digest(map[string]any{
		"partnership_id": i.PartnershipID, "agreement_version": i.AgreementVersion,
		"occurred_at": i.OccurredAt.UTC().Format(time.RFC3339Nano), "items": items,
	})
	if err != nil {
		return nil, fmt.Errorf("fingerprint drop request: %w", err)
	}
	return digest, nil
}

type AgreedProduct struct {
	ProductID string
	ShelfLife stock.ShelfLife
	Terms     pricing.Terms
}

type LastCount struct {
	At    time.Time
	Given bool
}

type Setting struct {
	LastCount        LastCount
	PartnershipID    string
	IsActive         bool
	AgreementID      string
	AgreementVersion int
	StoreID          string
	StoreName        string
	ConsignorID      string
	Products         map[string]AgreedProduct
}

type Draft struct {
	setting    Setting
	OccurredAt time.Time
	Items      []Item
}

func (s Setting) Check(input RecordInput, now time.Time) (Draft, error) {
	if !s.IsActive {
		return Draft{}, apperror.New(apperror.CodePartnershipNotActive)
	}
	if input.AgreementVersion != s.AgreementVersion {
		return Draft{}, apperror.New(apperror.CodeAgreementChanged).
			WithParam("n", strconv.Itoa(s.AgreementVersion))
	}
	occurredAt := input.OccurredAt.Truncate(time.Second)
	problems := field.Collect(findItemCountProblem(len(input.Items)), s.findDateProblem(occurredAt, now))
	items := make([]Item, 0, len(input.Items))
	for index, itemInput := range input.Items {
		item, itemProblems := s.checkItem(fmt.Sprintf("items[%d]", index), itemInput, occurredAt)
		problems = append(problems, itemProblems...)
		items = append(items, item)
	}
	if len(problems) > 0 {
		return Draft{}, apperror.Validation(problems...)
	}
	return Draft{setting: s, OccurredAt: occurredAt, Items: items}, nil
}

func findItemCountProblem(count int) field.Problem {
	if count < MinItems {
		return field.Invalid("items", apperror.FieldRequired)
	}
	if count > MaxItems {
		return field.Invalid("items", apperror.FieldTooMany, field.Limit("max", MaxItems))
	}
	return field.Problem{}
}

func (s Setting) findDateProblem(occurredAt, now time.Time) field.Problem {
	earliest := startOfDay(now).AddDate(0, 0, -maxDaysBack)
	if occurredAt.Before(earliest) || occurredAt.After(now.Add(futureTolerance)) {
		return field.Invalid("occurred_at", apperror.FieldOutOfRange)
	}
	if s.LastCount.Given && occurredAt.Before(s.LastCount.At.Truncate(time.Second)) {
		return field.Invalid("occurred_at", apperror.FieldBeforeLastDoc)
	}
	return field.Problem{}
}

func startOfDay(at time.Time) time.Time {
	local := at.In(period.Jakarta())
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location())
}

func (s Setting) checkItem(prefix string, input ItemInput, occurredAt time.Time) (Item, []apperror.FieldError) {
	product := s.Products[input.ProductID]
	expiry, expiryProblem := checkExpiry(input.ExpiryDate, occurredAt, product.ShelfLife)
	problems := field.Collect(
		s.findAgreedProblem(prefix, input.ProductID),
		findQtyProblem(prefix, input.Qty),
		prefixedProblem(prefix, expiryProblem),
	)
	return Item{ProductID: input.ProductID, Qty: input.Qty, Expiry: expiry, Terms: product.Terms}, problems
}

func (s Setting) findAgreedProblem(prefix, productID string) field.Problem {
	if _, isAgreed := s.Products[productID]; !isAgreed {
		return field.Invalid(prefix+".product_id", apperror.FieldNotInAgreement)
	}
	return field.Problem{}
}

func findQtyProblem(prefix string, qty int) field.Problem {
	if qty < MinQty || qty > MaxQty {
		return field.Invalid(prefix+".qty", apperror.FieldOutOfRange)
	}
	return field.Problem{}
}

type expiryProblem struct {
	code  apperror.FieldCode
	found bool
}

func prefixedProblem(prefix string, problem expiryProblem) field.Problem {
	if !problem.found {
		return field.Problem{}
	}
	return field.Invalid(prefix+".expiry_date", problem.code)
}

func checkExpiry(raw string, occurredAt time.Time, shelfLife stock.ShelfLife) (stock.Expiry, expiryProblem) {
	dropDay := startOfDay(occurredAt)
	if raw == "" {
		return defaultExpiry(dropDay, shelfLife), expiryProblem{}
	}
	date, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		return stock.NoExpiry(), expiryProblem{code: apperror.FieldInvalidDate, found: true}
	}
	if date.Before(time.Date(dropDay.Year(), dropDay.Month(), dropDay.Day(), 0, 0, 0, 0, time.UTC)) {
		return stock.NoExpiry(), expiryProblem{code: apperror.FieldExpiryBeforeDrop, found: true}
	}
	return stock.ExpiresOn(date), expiryProblem{}
}

func defaultExpiry(dropDay time.Time, shelfLife stock.ShelfLife) stock.Expiry {
	days, isKnown := shelfLife.Days()
	if !isKnown {
		return stock.NoExpiry()
	}
	return stock.ExpiresOn(dropDay.AddDate(0, 0, days))
}

type RecordIDs struct {
	DropID      string
	ItemIDs     []string
	MovementIDs []string
}

type Recording struct {
	IDs            RecordIDs
	Sequence       int
	RecordedBy     string
	IdempotencyKey string
	Fingerprint    []byte
	At             time.Time
}

func Number(at time.Time, sequence int) string {
	return fmt.Sprintf("%s-%s-%04d", NumberPrefix, period.MonthKey(at), sequence)
}

func (d Draft) Record(recording Recording) Drop {
	number := Number(recording.At, recording.Sequence)
	items := make([]Item, 0, len(d.Items))
	for index, item := range d.Items {
		item.ID = recording.IDs.ItemIDs[index]
		item.MovementID = recording.IDs.MovementIDs[index]
		item.BatchCode = batchCode(number, index+1)
		items = append(items, item)
	}
	return Drop{
		ID: recording.IDs.DropID, Number: number, PartnershipID: d.setting.PartnershipID,
		AgreementID: d.setting.AgreementID, AgreementVersion: d.setting.AgreementVersion,
		StoreID: d.setting.StoreID, ConsignorID: d.setting.ConsignorID, Status: Recorded,
		OccurredAt: d.OccurredAt, ReceivedAt: recording.At, RecordedBy: recording.RecordedBy,
		IdempotencyKey: recording.IdempotencyKey, Fingerprint: recording.Fingerprint, Items: items,
	}
}
