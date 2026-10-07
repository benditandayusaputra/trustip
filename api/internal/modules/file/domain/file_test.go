package domain

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

func fieldCode(t *testing.T, err error, field string) string {
	t.Helper()
	for _, problem := range apperror.From(err).Fields() {
		if problem.Field == field {
			return problem.Code
		}
	}
	t.Fatalf("err = %v has no %s field", err, field)
	return ""
}

func TestParsePurpose_NFR09_EachPurposeBelongsToTheRightOrganizationType(t *testing.T) {
	cases := []struct {
		raw     string
		orgType actor.OrganizationType
		valid   bool
	}{
		{"LOGO", actor.Store, true},
		{"LOGO", actor.Consignor, true},
		{"STORE_PHOTO", actor.Store, true},
		{"STORE_PHOTO", actor.Consignor, false},
		{"PAYMENT_QR", actor.Consignor, true},
		{"PAYMENT_QR", actor.Store, false},
		{"AVATAR", actor.Store, false},
		{"", actor.Store, false},
	}
	for _, tc := range cases {
		_, err := ParsePurpose(tc.raw, tc.orgType)

		if (err == nil) != tc.valid {
			t.Errorf("%s for %s: err = %v", tc.raw, tc.orgType, err)
		}
		if err != nil && fieldCode(t, err, "purpose") != string(apperror.FieldUnknownChoice) {
			t.Errorf("%s: wrong field code", tc.raw)
		}
	}
}

func TestVisibility_NFR09_PaymentQRIsPrivateAndProfileImagesArePublic(t *testing.T) {
	want := map[Purpose]Visibility{Logo: Public, StorePhoto: Public, PaymentQR: Private}
	for purpose, visibility := range want {
		if got := purpose.Visibility(); got != visibility {
			t.Errorf("%s = %s, want %s", purpose, got, visibility)
		}
	}
}

func TestCheckUploadSize_BRORG06_OneMegabyteBoundary(t *testing.T) {
	cases := map[int]bool{0: false, 1: true, MaxFileBytes: true, MaxFileBytes + 1: false}
	for size, valid := range cases {
		err := CheckUploadSize(size)

		if (err == nil) != valid {
			t.Errorf("%d bytes: err = %v", size, err)
		}
	}
}

func TestNewFile_KeepsTheRecodedImageFactsAndItsHash(t *testing.T) {
	image := Image{Data: []byte("jpeg"), ContentType: "image/jpeg", Extension: "jpg", Width: 640, Height: 480}
	now := time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)

	file := NewFile(Upload{ID: "file-1", OrganizationID: "org-1", CreatedBy: "user-1", Purpose: Logo},
		Stored{Image: image, Key: "abc.jpg"}, now)

	digest := sha256.Sum256([]byte("jpeg"))
	if file.Visibility != Public || file.SizeBytes != 4 || string(file.SHA256) != string(digest[:]) ||
		file.StorageKey != "abc.jpg" || file.Width != 640 || !file.CreatedAt.Equal(now) {
		t.Fatalf("file = %+v", file)
	}
}

func TestContentTypeOfKey_FollowsTheStoredExtension(t *testing.T) {
	cases := map[string]string{"a.jpg": "image/jpeg", "a.png": "image/png", "a.gif": ""}
	for key, want := range cases {
		if got := ContentTypeOfKey(key); got != want {
			t.Errorf("%s = %q", key, got)
		}
	}
}
