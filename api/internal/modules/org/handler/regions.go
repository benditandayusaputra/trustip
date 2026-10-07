package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type provinceResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type cityResponse struct {
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	ProvinceName string  `json:"province_name"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

func listProvinces(provinces ProvinceLister) fiber.Handler {
	return func(c fiber.Ctx) error {
		found, err := provinces.Execute(c.Context())
		if err != nil {
			return fmt.Errorf("list provinces: %w", err)
		}
		items := make([]provinceResponse, 0, len(found))
		for _, province := range found {
			items = append(items, provinceResponse{Code: province.Code, Name: province.Name})
		}
		return httpserver.RespondList(c, httpserver.Page[provinceResponse]{Items: items})
	}
}

func listCities(cities CityLister) fiber.Handler {
	return func(c fiber.Ctx) error {
		found, err := cities.Execute(c.Context(), c.Query("province_code"))
		if err != nil {
			return fmt.Errorf("list cities: %w", err)
		}
		return httpserver.RespondList(c, httpserver.Page[cityResponse]{Items: toCityResponses(found)})
	}
}

func toCityResponses(cities []domain.City) []cityResponse {
	items := make([]cityResponse, 0, len(cities))
	for _, city := range cities {
		items = append(items, cityResponse{
			Code: city.Code, Name: city.Name, ProvinceName: city.ProvinceName,
			Latitude: city.Centroid.Latitude, Longitude: city.Centroid.Longitude,
		})
	}
	return items
}
