package domain

import (
	"math"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
)

const (
	publicGridDegrees = 0.0045
	maxLatitude       = 90
	maxLongitude      = 180
	locationField     = "location"
)

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

func NewCoordinates(latitude, longitude float64) (Coordinates, error) {
	if math.Abs(latitude) > maxLatitude || math.Abs(longitude) > maxLongitude {
		return Coordinates{}, apperror.Validation(apperror.InvalidField(locationField, apperror.FieldOutOfRange))
	}
	return Coordinates{Latitude: latitude, Longitude: longitude}, nil
}

type Location struct {
	Exact  Coordinates
	Public Coordinates
}

type LocationInput struct {
	Given     bool
	Latitude  float64
	Longitude float64
}

func publicLocationRules() map[actor.OrganizationType]func(Coordinates) Coordinates {
	return map[actor.OrganizationType]func(Coordinates) Coordinates{
		actor.Store:     func(exact Coordinates) Coordinates { return exact },
		actor.Consignor: snapToPublicGrid,
	}
}

func PublicLocation(organizationType actor.OrganizationType, exact Coordinates) Coordinates {
	rule, known := publicLocationRules()[organizationType]
	if !known {
		return snapToPublicGrid(exact)
	}
	return rule(exact)
}

func snapToPublicGrid(exact Coordinates) Coordinates {
	return Coordinates{Latitude: snapToGrid(exact.Latitude), Longitude: snapToGrid(exact.Longitude)}
}

func snapToGrid(degrees float64) float64 {
	return math.Round(degrees/publicGridDegrees) * publicGridDegrees
}

func locate(organizationType actor.OrganizationType, input LocationInput) (Location, field.Problem) {
	exact, err := NewCoordinates(input.Latitude, input.Longitude)
	if err != nil {
		return Location{}, field.Invalid(locationField, apperror.FieldOutOfRange)
	}
	return Location{Exact: exact, Public: PublicLocation(organizationType, exact)}, field.Problem{}
}
