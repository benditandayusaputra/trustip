package domain

import (
	"slices"
	"strconv"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
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
	problems := collectProblems(
		findChoiceProblem("store_type", form.StoreType, StoreTypes()),
		findIDListProblem("accepted_category_ids", form.CategoryIDs, maxAcceptedCategories),
		findCommissionProblem(form.Commission),
		findLengthProblem("opening_hours", form.OpeningHours, lengthRange{max: MaxOpeningHoursLength}),
		findLengthProblem("other_requirements", form.Conditions.OtherRequirements,
			lengthRange{max: MaxOtherRequirementsSize}),
		findIDListProblem("photo_file_ids", form.PhotoFileIDs, MaxStorePhotos),
	)
	if len(problems) > 0 {
		return StoreProfile{}, apperror.Validation(problems...)
	}
	return StoreProfile(form), nil
}

func findChoiceProblem(field, value string, choices []string) problem {
	if value == "" || slices.Contains(choices, value) {
		return problem{}
	}
	return invalid(field, apperror.FieldUnknownChoice)
}

func findIDListProblem(field string, ids []string, limit int) problem {
	if len(ids) > limit {
		return invalid(field, apperror.FieldTooMany, apperror.Param{Name: "max", Value: strconv.Itoa(limit)})
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] || !fileIDFormat.MatchString(id) {
			return invalid(field, apperror.FieldUnknownChoice)
		}
		seen[id] = true
	}
	return problem{}
}

func findCommissionProblem(commission CommissionRange) problem {
	if !commission.Given {
		return problem{}
	}
	isInRange := commission.MinBPS >= 0 && commission.MaxBPS <= MaxCommissionBPS
	if isInRange && commission.MinBPS <= commission.MaxBPS {
		return problem{}
	}
	return invalid("commission_max_bps", apperror.FieldOutOfRange)
}
