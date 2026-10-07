package domain

import (
	"strconv"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
)

const (
	MaxStorePhotos           = 5
	MaxCommissionBPS         = 5000
	MaxOpeningHoursLength    = 200
	MaxOtherRequirementsSize = 500
	maxAcceptedCategories    = 12
)

func StoreTypes() []string {
	return []string{
		"GROCERY_STALL", "LOCAL_MINIMARKET", "SOUVENIR_SHOP", "CANTEEN", "CAFE",
		"BAKERY", "COOPERATIVE", "MARKET_KIOSK", "OTHER",
	}
}

type Acceptance struct {
	AcceptsFixedDeposit bool
	AcceptsCommission   bool
}

type CommissionRange struct {
	Given  bool
	MinBPS int
	MaxBPS int
}

type StoreConditions struct {
	RequiresDistributionPermit bool
	RequiresHalal              bool
	RequiresSample             bool
	OtherRequirements          string
}

type StoreProfileForm struct {
	StoreType   string
	CategoryIDs []string
	Acceptance
	Commission   CommissionRange
	Conditions   StoreConditions
	OpeningHours string
	PhotoFileIDs []string
}

type StoreProfile StoreProfileForm

type StoreProfileChange struct {
	OrganizationID string
	Profile        StoreProfile
	At             time.Time
}

func NewStoreProfile(form StoreProfileForm) (StoreProfile, error) {
	problems := field.Collect(
		field.CheckChoice("store_type", form.StoreType, StoreTypes()),
		findIDListProblem("accepted_category_ids", form.CategoryIDs, maxAcceptedCategories),
		findCommissionProblem(form.Commission),
		field.CheckLength("opening_hours", form.OpeningHours, field.Length{Max: MaxOpeningHoursLength}),
		field.CheckLength("other_requirements", form.Conditions.OtherRequirements,
			field.Length{Max: MaxOtherRequirementsSize}),
		findIDListProblem("photo_file_ids", form.PhotoFileIDs, MaxStorePhotos),
	)
	if len(problems) > 0 {
		return StoreProfile{}, apperror.Validation(problems...)
	}
	return StoreProfile(form), nil
}

func findIDListProblem(name string, ids []string, limit int) field.Problem {
	if len(ids) > limit {
		return field.Invalid(name, apperror.FieldTooMany, apperror.Param{Name: "max", Value: strconv.Itoa(limit)})
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] || !fileIDFormat.MatchString(id) {
			return field.Invalid(name, apperror.FieldUnknownChoice)
		}
		seen[id] = true
	}
	return field.Problem{}
}

func findCommissionProblem(commission CommissionRange) field.Problem {
	if !commission.Given {
		return field.Problem{}
	}
	isInRange := commission.MinBPS >= 0 && commission.MaxBPS <= MaxCommissionBPS
	if isInRange && commission.MinBPS <= commission.MaxBPS {
		return field.Problem{}
	}
	return field.Invalid("commission_max_bps", apperror.FieldOutOfRange)
}
