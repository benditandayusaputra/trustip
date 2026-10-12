package handler

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/modules/report/domain"
)

func TestProductPreview_UsesPublicPhotoAndEscapedDetailAddress(t *testing.T) {
	preview := toProductResponse(domain.ProductPreview{
		PhotoKey: "photo.webp", Slug: "roti-anton", SKU: "ROT/1",
	}, "https://files.trustip.id/")
	if preview == nil || preview.PhotoURL == nil || *preview.PhotoURL != "https://files.trustip.id/photo.webp" ||
		preview.Href != "/p/roti-anton/ROT%2F1" {
		t.Fatalf("preview = %+v", preview)
	}
}

func TestProductPreview_MissingPhotoKeepsDetailAvailable(t *testing.T) {
	preview := toProductResponse(domain.ProductPreview{Slug: "roti-anton", SKU: "ROT-1"}, "https://files.trustip.id")
	if preview == nil || preview.PhotoURL != nil || preview.Href != "/p/roti-anton/ROT-1" {
		t.Fatalf("preview = %+v", preview)
	}
}

func TestProductPreview_PartnerAndDailyRowsHaveNoProduct(t *testing.T) {
	if preview := toProductResponse(domain.ProductPreview{}, "https://files.trustip.id"); preview != nil {
		t.Fatalf("preview = %+v", preview)
	}
}
