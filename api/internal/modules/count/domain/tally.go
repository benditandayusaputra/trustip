package domain

import (
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/stock"
)

type ReturnReason string

const (
	ReasonUnsold     ReturnReason = "UNSOLD"
	ReasonExpired    ReturnReason = "EXPIRED"
	ReasonNearExpiry ReturnReason = "NEAR_EXPIRY"
	ReasonDamaged    ReturnReason = "DAMAGED"
	ReasonRecall     ReturnReason = "RECALL"
	ReasonOther      ReturnReason = "OTHER"
)

func ReturnReasons() []string {
	return []string{
		string(ReasonUnsold), string(ReasonExpired), string(ReasonNearExpiry),
		string(ReasonDamaged), string(ReasonRecall), string(ReasonOther),
	}
}

type PricedBatch struct {
	Stock stock.Batch
	Terms pricing.Terms
}

type LineInput struct {
	ProductID    string
	Batches      []PricedBatch
	CountedQty   int
	ReturnedQty  int
	ReturnReason string
}

type SaleAllocation struct {
	BatchID string
	Qty     int
	Split   pricing.Split
}

type Line struct {
	ProductID    string
	RecordedQty  int
	CountedQty   int
	SoldQty      int
	ReturnedQty  int
	KeptQty      int
	ReturnReason ReturnReason
	Sales        []SaleAllocation
	Returns      []stock.Allocation
	Split        pricing.Split
}

type Tally struct {
	Lines       []Line
	SoldQty     int
	ReturnedQty int
	Split       pricing.Split
}

func TallyCount(inputs []LineInput) (Tally, error) {
	tally := Tally{Lines: make([]Line, 0, len(inputs))}
	var problems []apperror.FieldError
	splits := make([]pricing.Split, 0, len(inputs))
	for index, input := range inputs {
		problems = append(problems, checkLine(fmt.Sprintf("lines[%d]", index), input)...)
	}
	if len(problems) > 0 {
		return Tally{}, apperror.Validation(problems...)
	}
	for _, input := range inputs {
		line, err := tallyLine(input)
		if err != nil {
			return Tally{}, fmt.Errorf("tally product %s: %w", input.ProductID, err)
		}
		tally.Lines = append(tally.Lines, line)
		tally.SoldQty += line.SoldQty
		tally.ReturnedQty += line.ReturnedQty
		splits = append(splits, line.Split)
	}
	tally.Split = pricing.Summarize(splits)
	return tally, nil
}

func checkLine(prefix string, input LineInput) []apperror.FieldError {
	return field.Collect(
		findCountedProblem(prefix, input.CountedQty, recordedQty(input.Batches)),
		findReturnedProblem(prefix, input.ReturnedQty, input.CountedQty),
		findReasonProblem(prefix, input.ReturnedQty, input.ReturnReason),
	)
}

func tallyLine(input LineInput) (Line, error) {
	recorded := recordedQty(input.Batches)
	line := Line{
		ProductID: input.ProductID, RecordedQty: recorded, CountedQty: input.CountedQty,
		SoldQty: recorded - input.CountedQty, ReturnedQty: input.ReturnedQty,
		KeptQty: input.CountedQty - input.ReturnedQty,
	}
	if line.ReturnedQty > 0 {
		line.ReturnReason = ReturnReason(input.ReturnReason)
	}
	return allocateLine(line, input.Batches)
}

func allocateLine(line Line, batches []PricedBatch) (Line, error) {
	stockBatches := make([]stock.Batch, 0, len(batches))
	termsByBatch := make(map[string]pricing.Terms, len(batches))
	for _, batch := range batches {
		stockBatches = append(stockBatches, batch.Stock)
		termsByBatch[batch.Stock.ID] = batch.Terms
	}
	sold, err := stock.Allocate(stockBatches, line.SoldQty)
	if err != nil {
		return Line{}, fmt.Errorf("allocate sold units: %w", err)
	}
	if line.Returns, err = stock.Allocate(stock.Deduct(stockBatches, sold), line.ReturnedQty); err != nil {
		return Line{}, fmt.Errorf("allocate returned units: %w", err)
	}
	line.Sales = make([]SaleAllocation, 0, len(sold))
	splits := make([]pricing.Split, 0, len(sold))
	for _, allocation := range sold {
		split := termsByBatch[allocation.BatchID].SplitOf(allocation.Qty)
		line.Sales = append(line.Sales, SaleAllocation{BatchID: allocation.BatchID, Qty: allocation.Qty, Split: split})
		splits = append(splits, split)
	}
	line.Split = pricing.Summarize(splits)
	return line, nil
}

func recordedQty(batches []PricedBatch) int {
	total := 0
	for _, batch := range batches {
		total += batch.Stock.Balance
	}
	return total
}

func findCountedProblem(prefix string, counted, recorded int) field.Problem {
	if counted < 0 {
		return field.Invalid(prefix+".counted_qty", apperror.FieldOutOfRange)
	}
	if counted > recorded {
		return field.Invalid(prefix+".counted_qty", apperror.FieldCountExceedsStock, field.Limit("n", recorded))
	}
	return field.Problem{}
}

func findReturnedProblem(prefix string, returned, counted int) field.Problem {
	if returned < 0 || returned > max(counted, 0) {
		return field.Invalid(prefix+".returned_qty", apperror.FieldOutOfRange)
	}
	return field.Problem{}
}

func findReasonProblem(prefix string, returned int, reason string) field.Problem {
	if returned <= 0 {
		return field.Problem{}
	}
	if reason == "" {
		return field.Invalid(prefix+".return_reason", apperror.FieldRequired)
	}
	return field.CheckChoice(prefix+".return_reason", reason, ReturnReasons())
}
