package domain

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/canonical"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/claimed"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
)

const (
	maxDaysBack     = 3
	futureTolerance = 5 * time.Minute
	MaxNoteLength   = 500
)

type Occasion string

const (
	StoreClosing Occasion = "STORE_CLOSING"
	Pickup       Occasion = "PICKUP"
	BeforeDrop   Occasion = "BEFORE_DROP"
	OtherTime    Occasion = "OTHER"
)

func Occasions() []string {
	return []string{string(StoreClosing), string(Pickup), string(BeforeDrop), string(OtherTime)}
}

type StockedProduct struct {
	ProductID string
	Name      string
	SKU       string
	Unit      string
	Batches   []PricedBatch
}

func (p StockedProduct) RecordedQty() int {
	return recordedQty(p.Batches)
}

type LastDocument struct {
	At    time.Time
	Given bool
}

type Settled struct {
	Until  time.Time
	Number string
	Given  bool
}

type Sheet struct {
	PartnershipID    string
	IsActive         bool
	AgreementVersion int
	StoreID          string
	StoreName        string
	ConsignorID      string
	ConsignorName    string
	LastDocument     LastDocument
	Settled          Settled
	Products         []StockedProduct
}

type RequestLine struct {
	ProductID    string
	RecordedQty  int
	CountedQty   int
	ReturnedQty  int
	ReturnReason string
}

type RecordInput struct {
	PartnershipID    string
	AgreementVersion int
	CountedAt        claimed.Time
	ClientRecordedAt claimed.Time
	RecordedOffline  bool
	Occasion         string
	Note             string
	Lines            []RequestLine
}

func (i RecordInput) Fingerprint() ([]byte, error) {
	lines := make([]any, 0, len(i.Lines))
	for _, line := range i.Lines {
		lines = append(lines, map[string]any{
			"product_id": line.ProductID, "recorded_qty": line.RecordedQty, "counted_qty": line.CountedQty,
			returnedQtyKey: line.ReturnedQty, "return_reason": line.ReturnReason,
		})
	}
	digest, err := canonical.Digest(map[string]any{
		"partnership_id": i.PartnershipID, "agreement_version": i.AgreementVersion, "occasion": i.Occasion,
		"counted_at": i.CountedAt.Text(), "client_recorded_at": i.ClientRecordedAt.Text(), "note": i.Note,
		"lines": lines,
	})
	if err != nil {
		return nil, fmt.Errorf("fingerprint count request: %w", err)
	}
	return digest, nil
}

type Draft struct {
	sheet            Sheet
	CountedAt        time.Time
	ClientRecordedAt time.Time
	RecordedOffline  bool
	Occasion         Occasion
	Note             string
	Tally            Tally
}

func (s Sheet) Check(input RecordInput, now time.Time) (Draft, error) {
	if !s.IsActive {
		return Draft{}, apperror.New(apperror.CodePartnershipNotActive)
	}
	if input.AgreementVersion != s.AgreementVersion {
		return Draft{}, apperror.New(apperror.CodeAgreementChanged).WithParam("n", strconv.Itoa(s.AgreementVersion))
	}
	lineInputs, problems := s.matchLines(input.Lines)
	clientRecordedAt := input.ClientRecordedAt.SettledBy(now).Truncate(time.Second)
	countedAt := input.CountedAt.Or(clientRecordedAt).Truncate(time.Second)
	problems = append(problems, field.Collect(
		s.findTimeProblem(countedAt, now), findOccasionProblem(input.Occasion),
		field.CheckLength("note", input.Note, field.Length{Max: MaxNoteLength}),
	)...)
	if len(problems) > 0 {
		return Draft{}, apperror.Validation(problems...)
	}
	if s.isStockChanged(input.Lines) {
		return Draft{}, apperror.New(apperror.CodeStockChanged)
	}
	tally, err := TallyCount(lineInputs)
	if err != nil {
		return Draft{}, fmt.Errorf("tally count: %w", err)
	}
	return Draft{
		sheet: s, CountedAt: countedAt, ClientRecordedAt: clientRecordedAt, RecordedOffline: input.RecordedOffline,
		Occasion: Occasion(input.Occasion), Note: input.Note, Tally: tally,
	}, nil
}

func (s Sheet) productOf(productID string) (StockedProduct, bool) {
	index := slices.IndexFunc(s.Products, func(product StockedProduct) bool { return product.ProductID == productID })
	if index < 0 {
		return StockedProduct{}, false
	}
	return s.Products[index], true
}

func (s Sheet) matchLines(lines []RequestLine) ([]LineInput, []apperror.FieldError) {
	problems := field.Collect(findLineCountProblem(len(lines)))
	inputs := make([]LineInput, 0, len(lines))
	seen := map[string]bool{}
	for index, line := range lines {
		product, isStocked := s.productOf(line.ProductID)
		if !isStocked || seen[line.ProductID] {
			problems = append(problems, apperror.InvalidField(
				fmt.Sprintf("lines[%d].product_id", index), apperror.FieldUnknownChoice))
		}
		seen[line.ProductID] = true
		inputs = append(inputs, LineInput{
			ProductID: line.ProductID, Batches: product.Batches, CountedQty: line.CountedQty,
			ReturnedQty: line.ReturnedQty, ReturnReason: line.ReturnReason,
		})
	}
	return inputs, problems
}

func findLineCountProblem(count int) field.Problem {
	if count == 0 {
		return field.Invalid("lines", apperror.FieldRequired)
	}
	return field.Problem{}
}

func (s Sheet) isStockChanged(lines []RequestLine) bool {
	return slices.ContainsFunc(lines, func(line RequestLine) bool {
		product, _ := s.productOf(line.ProductID)
		return product.RecordedQty() != line.RecordedQty
	})
}

func (s Sheet) findTimeProblem(countedAt, now time.Time) field.Problem {
	earliest := startOfDay(now).AddDate(0, 0, -maxDaysBack)
	if countedAt.Before(earliest) || countedAt.After(now.Add(futureTolerance)) {
		return field.Invalid("counted_at", apperror.FieldOutOfRange)
	}
	if s.LastDocument.Given && countedAt.Before(s.LastDocument.At.Truncate(time.Second)) {
		return field.Invalid("counted_at", apperror.FieldBeforeLastDoc)
	}
	if s.Settled.Given && countedAt.Before(s.Settled.Until) {
		return field.Invalid("counted_at", apperror.FieldPeriodSettled,
			apperror.Param{Name: "no", Value: s.Settled.Number})
	}
	return field.Problem{}
}

func startOfDay(at time.Time) time.Time {
	local := at.In(period.Jakarta())
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location())
}

func findOccasionProblem(occasion string) field.Problem {
	if occasion == "" {
		return field.Invalid("occasion", apperror.FieldRequired)
	}
	return field.CheckChoice("occasion", occasion, Occasions())
}
