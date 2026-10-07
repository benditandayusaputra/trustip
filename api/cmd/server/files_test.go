package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	filedomain "github.com/benditandayusaputra/trustip/api/internal/modules/file/domain"
	filerepository "github.com/benditandayusaputra/trustip/api/internal/modules/file/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver/edgetest"
	"github.com/benditandayusaputra/trustip/api/internal/platform/storage"
)

const (
	firstSequenceUserID = "00000000-0000-7000-8000-000000000001"
	testFilesBaseURL    = "http://localhost:8080/files"
)

type memoryFiles struct {
	mu    sync.Mutex
	files []filedomain.File
}

const exifWithGPS = "Exif\x00\x00GPSLatitude-7.0546"

func (m *memoryFiles) CreateFile(_ context.Context, file filedomain.File) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files = append(m.files, file)
	return nil
}

func (m *memoryFiles) GetFile(_ context.Context, organizationID, fileID string) (filedomain.File, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, file := range m.files {
		if file.ID == fileID && file.OrganizationID == organizationID {
			return file, nil
		}
	}
	return filedomain.File{}, apperror.NotFound()
}

type memoryStore struct {
	mu    sync.Mutex
	blobs map[string]memoryBlob
}

type memoryBlob struct {
	visibility filedomain.Visibility
	data       []byte
}

func (m *memoryStore) Put(
	_ context.Context, visibility filedomain.Visibility, data []byte, extension string,
) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%032x.%s", len(m.blobs)+1, extension)
	m.blobs[key] = memoryBlob{visibility: visibility, data: data}
	return key, nil
}

func (m *memoryStore) Read(visibility filedomain.Visibility, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	blob, ok := m.blobs[key]
	if !ok || blob.visibility != visibility {
		return nil, filedomain.ErrFileMissing
	}
	return blob.data, nil
}

func testFileComponents(now *clock.Manual) fileComponents {
	return fileComponents{
		store: &memoryStore{blobs: map[string]memoryBlob{}}, files: &memoryFiles{},
		clock: now, publicBaseURL: testFilesBaseURL,
	}
}

func databaseFileComponents(t *testing.T, pool *database.Pool, now *clock.Manual) fileComponents {
	t.Helper()
	return fileComponents{
		store: diskStore{disk: storage.NewDisk(t.TempDir())}, files: filerepository.NewFiles(pool),
		clock: now, publicBaseURL: testFilesBaseURL,
	}
}

func memberBrowser(t *testing.T, membership actor.Membership) *browser {
	t.Helper()
	now := clock.NewManual(testNow())
	auth := testAuthComponents(now)
	auth.memberships = memoryMemberships{firstSequenceUserID: membership}
	server := newTestServerWith(alwaysReady, now, auth)
	if got := perform(t, server, postUsers(t, registration("sari@contoh.id"))); got.status != http.StatusCreated {
		t.Fatalf("register = %d %s", got.status, got.body)
	}
	visitor := &browser{t: t, server: server, now: now}
	visitor.signIn()
	return visitor
}

const memberOrganizationID = "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c"

func storeOwner() actor.Membership {
	return actor.Membership{OrganizationID: memberOrganizationID, OrganizationType: actor.Store, Role: actor.Owner}
}

func photoWithLocation(t *testing.T) []byte {
	t.Helper()
	canvas := image.NewRGBA(image.Rect(0, 0, 32, 24))
	canvas.Set(1, 1, color.RGBA{R: 200, A: 255})
	var plain bytes.Buffer
	if err := jpeg.Encode(&plain, canvas, nil); err != nil {
		t.Fatalf("encode: %v", err)
	}
	segment := append([]byte{0xFF, 0xE1, 0x00, byte(len(exifWithGPS) + 2)}, exifWithGPS...)
	return append(append(append([]byte{}, plain.Bytes()[:2]...), segment...), plain.Bytes()[2:]...)
}

type part struct {
	name     string
	filename string
	data     []byte
}

func (b *browser) upload(parts ...part) exchange {
	b.t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, formPart := range parts {
		if err := writePart(writer, formPart); err != nil {
			b.t.Fatalf("write part: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		b.t.Fatalf("close form: %v", err)
	}
	req := httptest.NewRequestWithContext(b.t.Context(), http.MethodPost, "/api/v1/files", &body)
	req.Header.Set(fiber.HeaderContentType, writer.FormDataContentType())
	b.attachCookies(req)
	fromApp(req)
	b.visits++
	edgetest.SignFrom(b.t, req, edgetest.Origin{At: b.now.Now(), ClientIP: "198.51.100.210"})
	return b.send(req)
}

func writePart(writer *multipart.Writer, formPart part) error {
	if formPart.filename == "" {
		if err := writer.WriteField(formPart.name, string(formPart.data)); err != nil {
			return fmt.Errorf("write field %s: %w", formPart.name, err)
		}
		return nil
	}
	target, err := writer.CreateFormFile(formPart.name, formPart.filename)
	if err != nil {
		return fmt.Errorf("create file part: %w", err)
	}
	if _, err := target.Write(formPart.data); err != nil {
		return fmt.Errorf("write file part: %w", err)
	}
	return nil
}

func (b *browser) fetchPublic(path string) exchange {
	b.t.Helper()
	req := httptest.NewRequestWithContext(b.t.Context(), http.MethodGet, path, http.NoBody)
	return perform(b.t, b.server, req)
}

func TestUpload_RSEC70_PhotoIsRecodedAndServedPubliclyWithoutLocationData(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	got := visitor.upload(part{name: "purpose", data: []byte("STORE_PHOTO")},
		part{name: "file", filename: "rumah-saya.jpg", data: photoWithLocation(t)})

	if got.status != http.StatusCreated || strings.Contains(string(got.body), "rumah-saya") {
		t.Fatalf("upload = %d %s", got.status, got.body)
	}
	loadContract(t).requireConforms(t, got)
	key := quotedAfter(string(got.body), testFilesBaseURL+"/")
	served := visitor.fetchPublic("/files/" + key)
	if served.status != http.StatusOK || bytes.Contains(served.body, []byte("GPSLatitude")) {
		t.Fatalf("served = %d, %d bytes", served.status, len(served.body))
	}
	for header, want := range map[string]string{
		"Content-Type": "image/jpeg", "Content-Disposition": "inline", "X-Content-Type-Options": "nosniff",
		"Cross-Origin-Resource-Policy": "cross-origin", "Cache-Control": "public, max-age=31536000, immutable",
	} {
		if served.header.Get(header) != want {
			t.Errorf("%s = %q", header, served.header.Get(header))
		}
	}
}

func quotedAfter(text, start string) string {
	_, after, _ := strings.Cut(text, start)
	value, _, _ := strings.Cut(after, `"`)
	return value
}

func TestUpload_RSEC72_PaymentQRIsPrivateAndOnlyReadableThroughTheAPI(t *testing.T) {
	visitor := memberBrowser(t, actor.Membership{
		OrganizationID: memberOrganizationID, OrganizationType: actor.Consignor, Role: actor.Owner,
	})
	uploaded := visitor.upload(part{name: "purpose", data: []byte("PAYMENT_QR")},
		part{name: "file", filename: "qris.jpg", data: photoWithLocation(t)})
	fileURL := quotedAfter(string(uploaded.body), `"url":"`)

	own := visitor.get(fileURL)

	if !strings.HasPrefix(fileURL, "/api/v1/files/") || own.status != http.StatusOK ||
		own.header.Get("Cache-Control") != "no-store" || own.header.Get("Content-Type") != "image/jpeg" {
		t.Fatalf("url = %s, own = %d %v", fileURL, own.status, own.header)
	}
	loadContract(t).requireConforms(t, own)
}

func TestUpload_RSEC70_ContentThatIsNotAnImageIsRejected(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	got := visitor.upload(part{name: "purpose", data: []byte("LOGO")},
		part{name: "file", filename: "logo.jpg", data: []byte("<svg onload=alert(1)>")})

	requireCode(t, got, "VALIDATION_FAILED")
	if !strings.Contains(string(got.body), `"code":"INVALID_IMAGE"`) {
		t.Fatalf("body = %s", got.body)
	}
	loadContract(t).requireConforms(t, got)
}

func TestUpload_RRES05_UnknownFormPartIsRejected(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	got := visitor.upload(part{name: "purpose", data: []byte("LOGO")},
		part{name: "organization_id", data: []byte("org-2")},
		part{name: "file", filename: "logo.jpg", data: photoWithLocation(t)})

	requireCode(t, got, "UNKNOWN_FIELD")
}

func TestUpload_StaffCannotUploadProfileImages(t *testing.T) {
	visitor := memberBrowser(t, actor.Membership{
		OrganizationID: memberOrganizationID, OrganizationType: actor.Store, Role: actor.Staff,
	})

	got := visitor.upload(part{name: "purpose", data: []byte("LOGO")},
		part{name: "file", filename: "logo.jpg", data: photoWithLocation(t)})

	requireCode(t, got, "FORBIDDEN")
}

func TestReadFile_MalformedIDIsNotFound(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	requireCode(t, visitor.get("/api/v1/files/bukan-uuid"), "NOT_FOUND")
}

func TestPublicFile_UnknownKeyIsNotFound(t *testing.T) {
	visitor := memberBrowser(t, storeOwner())

	got := visitor.fetchPublic("/files/" + strings.Repeat("b2", 16) + ".jpg")

	requireCode(t, got, "NOT_FOUND")
}
