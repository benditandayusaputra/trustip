package domain

import (
	"slices"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
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
	problems := field.Collect(
		findRadiusProblem(form.DeliveryRadius),
		findChoicesProblem("delivery_days", form.DeliveryDays, DeliveryDays()),
		findChoicesProblem("preferred_store_types", form.PreferredStoreTypes, StoreTypes()),
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
	problems := field.Collect(
		field.CheckLength("payment_info", form.Info, field.Length{Max: MaxPaymentInfoLength}),
	)
	if form.QRFileID != "" && !fileIDFormat.MatchString(form.QRFileID) {
		problems = append(problems, apperror.InvalidField("payment_qr_file_id", apperror.FieldUnknownChoice))
	}
	if len(problems) > 0 {
		return PaymentInfo{}, apperror.Validation(problems...)
	}
	return PaymentInfo(form), nil
}
