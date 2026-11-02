package domain

import (
	"regexp"
	"slices"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/storetype"
)

const (
	MaxDeliveryRadiusKM    = 500
	MaxLegalNumberLength   = 50
	MaxPaymentInfoLength   = 500
	MaxBankNameLength      = 60
	MaxBankHolderLength    = 80
	MinAccountNumberLength = 5
	MaxAccountNumberLength = 30
	methodTransfer         = "TRANSFER"
	methodQRIS             = "QRIS"
)

var accountNumberFormat = regexp.MustCompile(`^[0-9][0-9 -]*$`)

func PaymentMethods() []string {
	return []string{"CASH", methodTransfer, methodQRIS}
}

func DeliveryDays() []string {
	return []string{"MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"}
}

type DeliveryRadius struct {
	Given      bool
	Kilometres int
}

type LegalNumbers struct {
	BusinessIDNo         string
	DistributionPermitNo string
	HalalCertNo          string
}

type ConsignorProfileForm struct {
	DeliveryRadius      DeliveryRadius
	DeliveryDays        []string
	PreferredStoreTypes []string
	Legal               LegalNumbers
	CanProvideSample    bool
}

type ConsignorProfile ConsignorProfileForm

type ConsignorProfileChange struct {
	OrganizationID string
	Profile        ConsignorProfile
	At             time.Time
}

func NewConsignorProfile(form ConsignorProfileForm) (ConsignorProfile, error) {
	problems := field.Collect(
		findRadiusProblem(form.DeliveryRadius),
		findChoicesProblem("delivery_days", form.DeliveryDays, DeliveryDays()),
		findChoicesProblem("preferred_store_types", form.PreferredStoreTypes, storetype.All()),
		field.CheckLength("business_id_no", form.Legal.BusinessIDNo, field.Length{Max: MaxLegalNumberLength}),
		field.CheckLength("distribution_permit_no", form.Legal.DistributionPermitNo,
			field.Length{Max: MaxLegalNumberLength}),
		field.CheckLength("halal_cert_no", form.Legal.HalalCertNo, field.Length{Max: MaxLegalNumberLength}),
	)
	if len(problems) > 0 {
		return ConsignorProfile{}, apperror.Validation(problems...)
	}
	return ConsignorProfile(form), nil
}

func findRadiusProblem(radius DeliveryRadius) field.Problem {
	if !radius.Given || (radius.Kilometres >= 1 && radius.Kilometres <= MaxDeliveryRadiusKM) {
		return field.Problem{}
	}
	return field.Invalid("delivery_radius_km", apperror.FieldOutOfRange)
}

func findChoicesProblem(name string, values, choices []string) field.Problem {
	seen := map[string]bool{}
	for _, value := range values {
		if seen[value] || !slices.Contains(choices, value) {
			return field.Invalid(name, apperror.FieldUnknownChoice)
		}
		seen[value] = true
	}
	return field.Problem{}
}

type BankAccount struct {
	Name   string
	Number string
	Holder string
}

type PaymentInfoForm struct {
	Info     string
	QRFileID string
	Methods  []string
	Bank     BankAccount
}

type PaymentInfo PaymentInfoForm

type PaymentInfoChange struct {
	OrganizationID string
	Payment        PaymentInfo
	At             time.Time
}

func NewPaymentInfo(form PaymentInfoForm) (PaymentInfo, error) {
	problems := field.Collect(append([]field.Problem{
		field.CheckLength("payment_info", form.Info, field.Length{Max: MaxPaymentInfoLength}),
		findMethodsProblem(form.Methods),
		findQRProblem(form),
		findAccountNumberProblem(form),
	}, findBankTextProblems(form)...)...)
	if len(problems) > 0 {
		return PaymentInfo{}, apperror.Validation(problems...)
	}
	return PaymentInfo(form), nil
}

func findMethodsProblem(methods []string) field.Problem {
	if len(methods) == 0 {
		return field.Invalid("payment_methods", apperror.FieldRequired)
	}
	return findChoicesProblem("payment_methods", methods, PaymentMethods())
}

func findQRProblem(form PaymentInfoForm) field.Problem {
	if form.QRFileID != "" && !fileIDFormat.MatchString(form.QRFileID) {
		return field.Invalid("payment_qr_file_id", apperror.FieldUnknownChoice)
	}
	if form.QRFileID == "" && slices.Contains(form.Methods, methodQRIS) {
		return field.Invalid("payment_qr_file_id", apperror.FieldRequired)
	}
	return field.Problem{}
}

func findBankTextProblems(form PaymentInfoForm) []field.Problem {
	texts := []struct {
		name    string
		value   string
		maximum int
	}{
		{"bank_name", form.Bank.Name, MaxBankNameLength},
		{"bank_account_holder", form.Bank.Holder, MaxBankHolderLength},
	}
	problems := make([]field.Problem, 0, len(texts))
	for _, text := range texts {
		if text.value == "" && slices.Contains(form.Methods, methodTransfer) {
			problems = append(problems, field.Invalid(text.name, apperror.FieldRequired))
			continue
		}
		problems = append(problems, field.CheckLength(text.name, text.value, field.Length{Max: text.maximum}))
	}
	return problems
}

func findAccountNumberProblem(form PaymentInfoForm) field.Problem {
	number := form.Bank.Number
	switch {
	case number == "" && slices.Contains(form.Methods, methodTransfer):
		return field.Invalid("bank_account_number", apperror.FieldRequired)
	case number == "":
		return field.Problem{}
	case !accountNumberFormat.MatchString(number):
		return field.Invalid("bank_account_number", apperror.FieldInvalidCharacters)
	}
	return field.CheckLength("bank_account_number", number,
		field.Length{Min: MinAccountNumberLength, Max: MaxAccountNumberLength})
}
