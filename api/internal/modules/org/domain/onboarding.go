package domain

import (
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/contact"
)

const (
	MinNameLength        = 2
	MaxNameLength        = 80
	MaxDescriptionLength = 1000
	MaxDistrictLength    = 80
	MaxAddressLength     = 200
)

var cityCodeFormat = regexp.MustCompile(`^[0-9]{2}\.[0-9]{2}$`)

type ProfileForm struct {
	Name        string
	Description string
	WhatsApp    string
	CityCode    string
	District    string
	Address     string
	Location    LocationInput
}

type OnboardingForm struct {
	Type string
	ProfileForm
}

type Profile struct {
	Name        string
	Description string
	WhatsApp    contact.WhatsAppNumber
	CityCode    string
	District    string
	Address     string
	Location    Location
	HasLocation bool
}

type Onboarding struct {
	Type actor.OrganizationType
	Profile
}

func NewOnboarding(form OnboardingForm) (Onboarding, error) {
	organizationType := actor.OrganizationType(form.Type)
	profile, problems := checkProfile(form.ProfileForm, organizationType)
	problems = append(collectProblems(findTypeProblem(form.Type)), problems...)
	if len(problems) > 0 {
		return Onboarding{}, apperror.Validation(problems...)
	}
	return Onboarding{Type: organizationType, Profile: profile}, nil
}

func checkProfile(form ProfileForm, organizationType actor.OrganizationType) (Profile, []apperror.FieldError) {
	whatsApp, whatsAppErr := contact.ParseWhatsAppNumber(form.WhatsApp)
	location, locationProblem := locateIfGiven(organizationType, form.Location)
	problems := collectProblems(
		findLengthProblem("name", form.Name, lengthRange{min: MinNameLength, max: MaxNameLength}),
		findLengthProblem("description", form.Description, lengthRange{max: MaxDescriptionLength}),
		findWhatsAppProblem(whatsAppErr),
		findCityProblem(form.CityCode),
		findLengthProblem("district", form.District, lengthRange{max: MaxDistrictLength}),
		findLengthProblem("address", form.Address, lengthRange{max: MaxAddressLength}),
		locationProblem,
	)
	return Profile{
		Name:        form.Name,
		Description: form.Description,
		WhatsApp:    whatsApp,
		CityCode:    form.CityCode,
		District:    form.District,
		Address:     form.Address,
		Location:    location,
		HasLocation: form.Location.Given,
	}, problems
}

func locateIfGiven(organizationType actor.OrganizationType, input LocationInput) (Location, problem) {
	if !input.Given {
		return Location{}, problem{}
	}
	return locate(organizationType, input)
}

type problem struct {
	field apperror.FieldError
	found bool
}

func collectProblems(candidates ...problem) []apperror.FieldError {
	var problems []apperror.FieldError
	for _, candidate := range candidates {
		if candidate.found {
			problems = append(problems, candidate.field)
		}
	}
	return problems
}

func invalid(field string, code apperror.FieldCode, params ...apperror.Param) problem {
	return problem{field: apperror.InvalidField(field, code, params...), found: true}
}

func findTypeProblem(organizationType string) problem {
	if organizationType == string(actor.Consignor) || organizationType == string(actor.Store) {
		return problem{}
	}
	return invalid("type", apperror.FieldUnknownChoice)
}

func findWhatsAppProblem(err error) problem {
	if err == nil {
		return problem{}
	}
	return invalid("whatsapp", apperror.FieldInvalidWhatsApp)
}

func findCityProblem(cityCode string) problem {
	if cityCode == "" {
		return invalid("city_code", apperror.FieldRequired)
	}
	if !cityCodeFormat.MatchString(cityCode) {
		return invalid("city_code", apperror.FieldUnknownChoice)
	}
	return problem{}
}

type lengthRange struct {
	min int
	max int
}

func findLengthProblem(field, value string, allowed lengthRange) problem {
	length := utf8.RuneCountInString(value)
	if length < allowed.min {
		return invalid(field, apperror.FieldTooShort, limitParam("min", allowed.min))
	}
	if length > allowed.max {
		return invalid(field, apperror.FieldTooLong, limitParam("max", allowed.max))
	}
	return problem{}
}

func limitParam(name string, limit int) apperror.Param {
	return apperror.Param{Name: name, Value: strconv.Itoa(limit)}
}

type FoundingIDs struct {
	OrganizationID string
	MembershipID   string
	FounderID      string
}

type Organization struct {
	ID string
	Onboarding
	Slug      Slug
	CityName  string
	CreatedAt time.Time
}

type Membership struct {
	ID     string
	UserID string
	Role   actor.Role
}

type Founding struct {
	Organization Organization
	Owner        Membership
}

func (o Onboarding) Found(ids FoundingIDs, slug Slug, now time.Time) Founding {
	return Founding{
		Organization: Organization{ID: ids.OrganizationID, Onboarding: o, Slug: slug, CreatedAt: now},
		Owner:        Membership{ID: ids.MembershipID, UserID: ids.FounderID, Role: actor.Owner},
	}
}
