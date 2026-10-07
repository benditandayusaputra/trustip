package domain

import (
	"regexp"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/contact"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
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
	problems = append(field.Collect(findTypeProblem(form.Type)), problems...)
	if len(problems) > 0 {
		return Onboarding{}, apperror.Validation(problems...)
	}
	return Onboarding{Type: organizationType, Profile: profile}, nil
}

func checkProfile(form ProfileForm, organizationType actor.OrganizationType) (Profile, []apperror.FieldError) {
	whatsApp, whatsAppErr := contact.ParseWhatsAppNumber(form.WhatsApp)
	location, locationProblem := locateIfGiven(organizationType, form.Location)
	problems := field.Collect(
		field.CheckLength("name", form.Name, field.Length{Min: MinNameLength, Max: MaxNameLength}),
		field.CheckLength("description", form.Description, field.Length{Max: MaxDescriptionLength}),
		findWhatsAppProblem(whatsAppErr),
		findCityProblem(form.CityCode),
		field.CheckLength("district", form.District, field.Length{Max: MaxDistrictLength}),
		field.CheckLength("address", form.Address, field.Length{Max: MaxAddressLength}),
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

func locateIfGiven(organizationType actor.OrganizationType, input LocationInput) (Location, field.Problem) {
	if !input.Given {
		return Location{}, field.Problem{}
	}
	return locate(organizationType, input)
}

func findTypeProblem(organizationType string) field.Problem {
	if organizationType == string(actor.Consignor) || organizationType == string(actor.Store) {
		return field.Problem{}
	}
	return field.Invalid("type", apperror.FieldUnknownChoice)
}

func findWhatsAppProblem(err error) field.Problem {
	if err == nil {
		return field.Problem{}
	}
	return field.Invalid("whatsapp", apperror.FieldInvalidWhatsApp)
}

func findCityProblem(cityCode string) field.Problem {
	if cityCode == "" {
		return field.Invalid("city_code", apperror.FieldRequired)
	}
	if !cityCodeFormat.MatchString(cityCode) {
		return field.Invalid("city_code", apperror.FieldUnknownChoice)
	}
	return field.Problem{}
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
