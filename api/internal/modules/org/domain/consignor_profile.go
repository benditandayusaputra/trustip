package domain

import (
	"slices"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const (
	MaxDeliveryRadiusKM  = 500
	MaxLegalNumberLength = 50
	MaxPaymentInfoLength = 500
)

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
	problems := collectProblems(
		findRadiusProblem(form.DeliveryRadius),
		findChoicesProblem("delivery_days", form.DeliveryDays, DeliveryDays()),
		findChoicesProblem("preferred_store_types", form.PreferredStoreTypes, StoreTypes()),
		findLengthProblem("business_id_no", form.Legal.BusinessIDNo, lengthRange{max: MaxLegalNumberLength}),
		findLengthProblem("distribution_permit_no", form.Legal.DistributionPermitNo,
			lengthRange{max: MaxLegalNumberLength}),
		findLengthProblem("halal_cert_no", form.Legal.HalalCertNo, lengthRange{max: MaxLegalNumberLength}),
	)
	if len(problems) > 0 {
		return ConsignorProfile{}, apperror.Validation(problems...)
	}
	return ConsignorProfile(form), nil
}

func findRadiusProblem(radius DeliveryRadius) problem {
	if !radius.Given || (radius.Kilometres >= 1 && radius.Kilometres <= MaxDeliveryRadiusKM) {
		return problem{}
	}
	return invalid("delivery_radius_km", apperror.FieldOutOfRange)
}

func findChoicesProblem(field string, values, choices []string) problem {
	seen := map[string]bool{}
	for _, value := range values {
		if seen[value] || !slices.Contains(choices, value) {
			return invalid(field, apperror.FieldUnknownChoice)
		}
		seen[value] = true
	}
	return problem{}
}

type PaymentInfoForm struct {
	Info     string
	QRFileID string
}

type PaymentInfo PaymentInfoForm

type PaymentInfoChange struct {
	OrganizationID string
	Payment        PaymentInfo
	At             time.Time
}

func NewPaymentInfo(form PaymentInfoForm) (PaymentInfo, error) {
	problems := collectProblems(
		findLengthProblem("payment_info", form.Info, lengthRange{max: MaxPaymentInfoLength}),
	)
	if form.QRFileID != "" && !fileIDFormat.MatchString(form.QRFileID) {
		problems = append(problems, apperror.InvalidField("payment_qr_file_id", apperror.FieldUnknownChoice))
	}
	if len(problems) > 0 {
		return PaymentInfo{}, apperror.Validation(problems...)
	}
	return PaymentInfo(form), nil
}
