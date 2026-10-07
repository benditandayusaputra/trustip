package pricing

import (
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/money"
)

const (
	MaxCommissionBPS = 5000
	ExampleQuantity  = 10
	wholeBPS         = 10000
	halfBPS          = wholeBPS / 2
)

type Model string

const (
	FixedDeposit Model = "FIXED_DEPOSIT"
	Commission   Model = "COMMISSION"
)

func Models() []string {
	return []string{string(FixedDeposit), string(Commission)}
}

type TermsInput struct {
	Model         string
	RetailPrice   int64
	CommissionBPS int
	DepositPrice  int64
}

type Terms struct {
	Model         Model
	RetailPrice   money.Rupiah
	CommissionBPS int
	DepositPrice  money.Rupiah
}

type Split struct {
	Gross          money.Rupiah
	StoreShare     money.Rupiah
	ConsignorShare money.Rupiah
}

type shareRule func(terms Terms, qty int64, gross money.Rupiah) Split

func shareRules() map[Model]shareRule {
	return map[Model]shareRule{
		Commission:   commissionShare,
		FixedDeposit: fixedDepositShare,
	}
}

func commissionShare(terms Terms, _ int64, gross money.Rupiah) Split {
	storeShare := roundHalfUp(gross.Int64() * int64(terms.CommissionBPS))
	return Split{Gross: gross, StoreShare: storeShare, ConsignorShare: gross - storeShare}
}

func fixedDepositShare(terms Terms, qty int64, gross money.Rupiah) Split {
	consignorShare := money.Rupiah(qty * terms.DepositPrice.Int64())
	return Split{Gross: gross, StoreShare: gross - consignorShare, ConsignorShare: consignorShare}
}

func roundHalfUp(basisPointRupiah int64) money.Rupiah {
	return money.Rupiah((basisPointRupiah + halfBPS) / wholeBPS)
}

func (t Terms) SplitOf(qty int) Split {
	units := int64(qty)
	gross := money.Rupiah(units * t.RetailPrice.Int64())
	return shareRules()[t.Model](t, units, gross)
}

func (t Terms) PerUnit() Split {
	return t.SplitOf(1)
}

func (t Terms) Example() Split {
	return t.SplitOf(ExampleQuantity)
}

func Summarize(splits []Split) Split {
	var total Split
	for _, split := range splits {
		total.Gross += split.Gross
		total.StoreShare += split.StoreShare
		total.ConsignorShare += split.ConsignorShare
	}
	return total
}

func NewTerms(input TermsInput) (Terms, error) {
	model := Model(input.Model)
	retail, retailErr := money.NewRupiah(input.RetailPrice)
	problems := field.Collect(
		findModelProblem(input.Model),
		findRetailProblem(retailErr),
	)
	terms := Terms{Model: model, RetailPrice: retail}
	switch model {
	case Commission:
		terms.CommissionBPS = input.CommissionBPS
		problems = append(problems, field.Collect(findCommissionProblem(input.CommissionBPS))...)
	case FixedDeposit:
		deposit, depositProblem := checkDeposit(input.DepositPrice, retail)
		terms.DepositPrice = deposit
		problems = append(problems, field.Collect(depositProblem)...)
	}
	if len(problems) > 0 {
		return Terms{}, apperror.Validation(problems...)
	}
	return terms, nil
}

func findModelProblem(model string) field.Problem {
	if model == "" {
		return field.Invalid("pricing_model", apperror.FieldRequired)
	}
	return field.CheckChoice("pricing_model", model, Models())
}

func findRetailProblem(err error) field.Problem {
	if err != nil {
		return field.Invalid("retail_price", apperror.FieldOutOfRange)
	}
	return field.Problem{}
}

func findCommissionProblem(commissionBPS int) field.Problem {
	if commissionBPS < 0 || commissionBPS > MaxCommissionBPS {
		return field.Invalid("commission_bps", apperror.FieldOutOfRange)
	}
	return field.Problem{}
}

func checkDeposit(amount int64, retail money.Rupiah) (money.Rupiah, field.Problem) {
	deposit, err := money.NewRupiah(amount)
	if err != nil {
		return 0, field.Invalid("deposit_price", apperror.FieldOutOfRange)
	}
	if deposit > retail {
		return 0, field.Invalid("deposit_price", apperror.FieldExceedsRetail)
	}
	return deposit, field.Problem{}
}
