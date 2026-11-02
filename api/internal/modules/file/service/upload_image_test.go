package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/file/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/file/service"
)

type fakeRecoder struct {
	err   error
	image domain.Image
}

func (f fakeRecoder) Recode([]byte) (domain.Image, error) {
	return f.image, f.err
}

type storedBlob struct {
	visibility domain.Visibility
	data       []byte
}

type fakeStore struct{ blobs map[string]storedBlob }

func (f *fakeStore) Put(
	_ context.Context, visibility domain.Visibility, data []byte, extension string,
) (string, error) {
	key := strings.Repeat("a1", 16) + "." + extension
	f.blobs[key] = storedBlob{visibility: visibility, data: data}
	return key, nil
}

func (f *fakeStore) Read(visibility domain.Visibility, key string) ([]byte, error) {
	blob, ok := f.blobs[key]
	if !ok || blob.visibility != visibility {
		return nil, domain.ErrFileMissing
	}
	return blob.data, nil
}

type fakeFiles struct{ created []domain.File }

func (f *fakeFiles) CreateFile(_ context.Context, file domain.File) error {
	f.created = append(f.created, file)
	return nil
}

func (f *fakeFiles) GetFile(_ context.Context, organizationID, fileID string) (domain.File, error) {
	for _, file := range f.created {
		if file.ID == fileID && file.OrganizationID == organizationID {
			return file, nil
		}
	}
	return domain.File{}, apperror.NotFound()
}

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC) }

type fixedIDs struct{}

func (fixedIDs) NewID() (string, error) { return "file-1", nil }

type uploadFixture struct {
	store   *fakeStore
	files   *fakeFiles
	recoder fakeRecoder
}

func newUploadFixture() *uploadFixture {
	return &uploadFixture{
		store: &fakeStore{blobs: map[string]storedBlob{}},
		files: &fakeFiles{},
		recoder: fakeRecoder{image: domain.Image{
			Data: []byte("recoded"), ContentType: "image/png", Extension: "png", Width: 300, Height: 300,
		}},
	}
}

func (f *uploadFixture) useCase() *service.UploadImage {
	return service.NewUploadImage(service.UploadDeps{
		Images: f.recoder, Store: f.store, Files: f.files, Clock: fixedClock{}, IDs: fixedIDs{},
	})
}

func member(orgType actor.OrganizationType, role actor.Role) actor.Actor {
	return actor.Actor{UserID: "user-1", Membership: actor.Membership{
		OrganizationID: "org-1", OrganizationType: orgType, Role: role,
	}}
}

func upload(purpose string) service.UploadCommand {
	return service.UploadCommand{Purpose: purpose, Data: []byte("original-with-exif")}
}

func TestUploadImage_StaffCannotChangeTheProfile(t *testing.T) {
	fixture := newUploadFixture()

	_, err := fixture.useCase().Execute(t.Context(), member(actor.Consignor, actor.Staff), upload("PAYMENT_QR"))

	if apperror.From(err).Code() != apperror.CodeForbidden || len(fixture.files.created) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestUploadImage_PurposeOfTheOtherOrganizationTypeIsRejected(t *testing.T) {
	_, err := newUploadFixture().useCase().Execute(t.Context(), member(actor.Store, actor.Owner), upload("PAYMENT_QR"))

	if apperror.From(err).Code() != apperror.CodeValidationFailed {
		t.Fatalf("err = %v", err)
	}
}

func TestUploadImage_BRORG06_FileOverOneMegabyteIsRejectedBeforeDecoding(t *testing.T) {
	fixture := newUploadFixture()
	command := upload("LOGO")
	command.Data = make([]byte, domain.MaxFileBytes+1)

	_, err := fixture.useCase().Execute(t.Context(), member(actor.Store, actor.Owner), command)

	if !hasFileCode(err, "FILE_TOO_LARGE") {
		t.Fatalf("err = %v", err)
	}
}

func TestUploadImage_RSEC70_ContentThatIsNotAnImageIsRejected(t *testing.T) {
	for cause, code := range map[error]string{
		domain.ErrUnsupportedImage: "INVALID_IMAGE",
		domain.ErrImageTooLarge:    "IMAGE_TOO_LARGE",
	} {
		fixture := newUploadFixture()
		fixture.recoder.err = cause

		_, err := fixture.useCase().Execute(t.Context(), member(actor.Store, actor.Owner), upload("LOGO"))

		if !hasFileCode(err, code) || len(fixture.store.blobs) != 0 {
			t.Errorf("%v: err = %v", cause, err)
		}
	}
}

func TestUploadImage_RecodedImageThatGrewPastTheLimitIsRejected(t *testing.T) {
	fixture := newUploadFixture()
	fixture.recoder.image.Data = make([]byte, domain.MaxFileBytes+1)

	_, err := fixture.useCase().Execute(t.Context(), member(actor.Store, actor.Owner), upload("LOGO"))

	if !hasFileCode(err, "FILE_TOO_LARGE") || len(fixture.store.blobs) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestUploadImage_RSEC54_StoresOnlyTheRecodedBytesUnderAServerKey(t *testing.T) {
	fixture := newUploadFixture()

	file, err := fixture.useCase().Execute(t.Context(), member(actor.Consignor, actor.Owner), upload("PAYMENT_QR"))
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	blob := fixture.store.blobs[file.StorageKey]
	if string(blob.data) != "recoded" || blob.visibility != domain.Private || file.OrganizationID != "org-1" ||
		len(fixture.files.created) != 1 {
		t.Fatalf("file = %+v, blob = %+v", file, blob)
	}
}

func TestUploadImage_PaymentProofMayBeAPlainPDF(t *testing.T) {
	fixture := newUploadFixture()
	receipt := []byte("%PDF-1.4\n1 0 obj << /Type /Catalog >> endobj\ntrailer << /Root 1 0 R >>\n%%EOF\n")

	file, err := fixture.useCase().Execute(t.Context(), member(actor.Store, actor.Owner),
		service.UploadCommand{Purpose: "PAYMENT_PROOF", Data: receipt})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	blob := fixture.store.blobs[file.StorageKey]
	if file.ContentType != "application/pdf" || !strings.HasSuffix(file.StorageKey, ".pdf") ||
		string(blob.data) != string(receipt) || blob.visibility != domain.Private || file.Width != 0 {
		t.Fatalf("file = %+v", file)
	}
}

func TestUploadImage_RSEC70_PDFIsOnlyForPaymentProofAndMustBeSafe(t *testing.T) {
	scripted := []byte("%PDF-1.4\n1 0 obj << /S /JavaScript /JS (x) >> endobj\n%%EOF\n")
	cases := []struct {
		purpose string
		orgType actor.OrganizationType
		data    []byte
		code    string
	}{
		{"LOGO", actor.Store, []byte("%PDF-1.4\n%%EOF\n"), "INVALID_IMAGE"},
		{"PAYMENT_PROOF", actor.Store, scripted, "INVALID_PROOF"},
		{"PAYMENT_PROOF", actor.Store, []byte("plain text"), "INVALID_PROOF"},
	}
	for _, tc := range cases {
		fixture := newUploadFixture()
		fixture.recoder.err = domain.ErrUnsupportedImage

		_, err := fixture.useCase().Execute(t.Context(), member(tc.orgType, actor.Owner),
			service.UploadCommand{Purpose: tc.purpose, Data: tc.data})

		if !hasFileCode(err, tc.code) || len(fixture.store.blobs) != 0 {
			t.Errorf("%s %q: err = %v", tc.purpose, tc.data, err)
		}
	}
}

func hasFileCode(err error, code string) bool {
	for _, problem := range apperror.From(err).Fields() {
		if problem.Field == "file" && problem.Code == code {
			return true
		}
	}
	return false
}

func TestReadFile_BRAUTHZ02_FileOfAnotherOrganizationIsNotFound(t *testing.T) {
	fixture := newUploadFixture()
	if _, err := fixture.useCase().Execute(t.Context(), member(actor.Store, actor.Owner), upload("LOGO")); err != nil {
		t.Fatalf("upload: %v", err)
	}
	stranger := member(actor.Store, actor.Owner)
	stranger.Membership.OrganizationID = "org-2"

	_, err := service.NewReadFile(fixture.files, fixture.store).Execute(t.Context(), stranger, "file-1")

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestReadFile_RSEC72_MemberReadsItsOwnFile(t *testing.T) {
	fixture := newUploadFixture()
	consignor := member(actor.Consignor, actor.Owner)
	if _, err := fixture.useCase().Execute(t.Context(), consignor, upload("PAYMENT_QR")); err != nil {
		t.Fatalf("upload: %v", err)
	}

	content, err := service.NewReadFile(fixture.files, fixture.store).
		Execute(t.Context(), member(actor.Consignor, actor.Staff), "file-1")

	if err != nil || string(content.Data) != "recoded" || content.ContentType != "image/png" {
		t.Fatalf("content = %+v, err = %v", content, err)
	}
}

func TestReadPublicFile_PrivateKeyIsNotServedPublicly(t *testing.T) {
	fixture := newUploadFixture()
	file, err := fixture.useCase().Execute(t.Context(), member(actor.Consignor, actor.Owner), upload("PAYMENT_QR"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	_, err = service.NewReadPublicFile(fixture.store).Execute(t.Context(), file.StorageKey)

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestReadPublicFile_PublicKeyIsServedWithItsType(t *testing.T) {
	fixture := newUploadFixture()
	file, err := fixture.useCase().Execute(t.Context(), member(actor.Store, actor.Owner), upload("STORE_PHOTO"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	content, err := service.NewReadPublicFile(fixture.store).Execute(t.Context(), file.StorageKey)

	if err != nil || content.ContentType != "image/png" {
		t.Fatalf("content = %+v, err = %v", content, err)
	}
}
