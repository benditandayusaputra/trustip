package domain

import (
	"regexp"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

var fileIDFormat = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type ProfileUpdateForm struct {
	ProfileForm
	LogoFileID string
}

type ProfileUpdate struct {
	Profile
	LogoFileID string
}

type ProfileChange struct {
	OrganizationID string
	Update         ProfileUpdate
	At             time.Time
}

func NewProfileUpdate(form ProfileUpdateForm, organizationType actor.OrganizationType) (ProfileUpdate, error) {
	profile, problems := checkProfile(form.ProfileForm, organizationType)
	if form.LogoFileID != "" && !fileIDFormat.MatchString(form.LogoFileID) {
		problems = append(problems, apperror.InvalidField("logo_file_id", apperror.FieldUnknownChoice))
	}
	if len(problems) > 0 {
		return ProfileUpdate{}, apperror.Validation(problems...)
	}
	return ProfileUpdate{Profile: profile, LogoFileID: form.LogoFileID}, nil
}

type FileRef struct {
	ID         string
	StorageKey string
}

type Settings struct {
	ID       string
	Type     actor.OrganizationType
	Slug     Slug
	Profile  Profile
	CityName string
	Logo     FileRef
}
