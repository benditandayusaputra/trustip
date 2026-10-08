package domain

import (
	"math"
	"regexp"
	"strconv"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/field"
)

const (
	maxLatitude  = 90
	maxLongitude = 180
	originField  = "origin"
	cityField    = "city_code"
)

var cityCodeFormat = regexp.MustCompile(`^[0-9]{2}\.[0-9]{2}$`)

type OriginKind string

const (
	OwnLocation     OriginKind = "ORGANIZATION"
	CurrentPosition OriginKind = "HERE"
	CityCentre      OriginKind = "CITY"
)

type Point struct {
	Latitude  float64
	Longitude float64
}

type OriginForm struct {
	Kind      string
	Latitude  string
	Longitude string
	CityCode  string
}

type OriginRequest struct {
	Kind     OriginKind
	Point    Point
	CityCode string
}

type Origin struct {
	Kind    OriginKind
	Point   Point
	IsKnown bool
}

type ConsignorFacts struct {
	DeliveryRadiusKm    int
	PreferredStoreTypes []string
	Categories          []Category
}

type Viewer struct {
	OrganizationID string
	Type           actor.OrganizationType
	Location       Point
	HasLocation    bool
	City           Point
	HasCity        bool
	Consignor      ConsignorFacts
}

func readOrigin(form OriginForm) (OriginRequest, []apperror.FieldError) {
	switch OriginKind(form.Kind) {
	case "", OwnLocation:
		return OriginRequest{Kind: OwnLocation}, nil
	case CurrentPosition:
		point, ok := readPoint(form.Latitude, form.Longitude)
		if !ok {
			return OriginRequest{}, field.Collect(field.Invalid(originField, apperror.FieldOutOfRange))
		}
		return OriginRequest{Kind: CurrentPosition, Point: point}, nil
	case CityCentre:
		if !cityCodeFormat.MatchString(form.CityCode) {
			return OriginRequest{}, field.Collect(field.Invalid(cityField, apperror.FieldUnknownChoice))
		}
		return OriginRequest{Kind: CityCentre, CityCode: form.CityCode}, nil
	}
	return OriginRequest{}, field.Collect(field.Invalid(originField, apperror.FieldUnknownChoice))
}

func readPoint(rawLatitude, rawLongitude string) (Point, bool) {
	latitude, latitudeErr := strconv.ParseFloat(rawLatitude, 64)
	longitude, longitudeErr := strconv.ParseFloat(rawLongitude, 64)
	if latitudeErr != nil || longitudeErr != nil {
		return Point{}, false
	}
	if math.Abs(latitude) > maxLatitude || math.Abs(longitude) > maxLongitude {
		return Point{}, false
	}
	return Point{Latitude: latitude, Longitude: longitude}, true
}

func (r OriginRequest) ResolveFor(viewer Viewer) (Origin, error) {
	switch r.Kind {
	case CurrentPosition:
		return Origin{Kind: r.Kind, Point: r.Point, IsKnown: true}, nil
	case CityCentre:
		if !viewer.HasCity {
			return Origin{}, apperror.Validation(apperror.InvalidField(cityField, apperror.FieldUnknownChoice))
		}
		return Origin{Kind: r.Kind, Point: viewer.City, IsKnown: true}, nil
	case OwnLocation:
	}
	return Origin{Kind: OwnLocation, Point: viewer.Location, IsKnown: viewer.HasLocation}, nil
}

func (v Viewer) DefaultCategoryIDs() []string {
	ids := make([]string, 0, len(v.Consignor.Categories))
	for _, category := range v.Consignor.Categories {
		ids = append(ids, category.ID)
	}
	return ids
}

func (v Viewer) IsConsignor() bool {
	return v.Type == actor.Consignor
}
