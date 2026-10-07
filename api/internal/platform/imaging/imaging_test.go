package imaging_test

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/platform/imaging"
)

const exifWithGPS = "Exif\x00\x00GPSLatitude-7.0546"

func photo(t *testing.T, width, height int) []byte {
	t.Helper()
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := range width {
		canvas.Set(x, 0, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	}
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, canvas, nil); err != nil {
		t.Fatalf("encode: %v", err)
	}
	return encoded.Bytes()
}

func withExif(t *testing.T, plain []byte) []byte {
	t.Helper()
	segment := append([]byte{0xFF, 0xE1, 0x00, byte(len(exifWithGPS) + 2)}, exifWithGPS...)
	return append(append(append([]byte{}, plain[:2]...), segment...), plain[2:]...)
}

func TestRecode_RSEC70_JPEGIsReencodedWithoutMetadata(t *testing.T) {
	tagged := withExif(t, photo(t, 40, 30))
	if !bytes.Contains(tagged, []byte("GPSLatitude")) {
		t.Fatal("fixture must carry location metadata")
	}

	recoded, err := imaging.Recode(tagged)
	if err != nil {
		t.Fatalf("recode: %v", err)
	}

	if bytes.Contains(recoded.Data, []byte("GPSLatitude")) || bytes.Contains(recoded.Data, []byte("Exif")) {
		t.Fatal("recoded image still carries metadata")
	}
	if recoded.ContentType != "image/jpeg" || recoded.Extension != "jpg" ||
		recoded.Width != 40 || recoded.Height != 30 {
		t.Fatalf("recoded = %s %s %dx%d", recoded.ContentType, recoded.Extension, recoded.Width, recoded.Height)
	}
}

func TestRecode_PNGStaysPNGSoTransparencySurvives(t *testing.T) {
	canvas := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	canvas.Set(1, 1, color.NRGBA{A: 0})
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, canvas); err != nil {
		t.Fatalf("encode: %v", err)
	}

	recoded, err := imaging.Recode(encoded.Bytes())

	if err != nil || recoded.ContentType != "image/png" || recoded.Extension != "png" {
		t.Fatalf("recoded = %+v, err = %v", recoded.ContentType, err)
	}
}

func TestRecode_RSEC70_TypeComesFromTheContentNotTheName(t *testing.T) {
	for name, data := range map[string][]byte{
		"script":    []byte("<script>alert(1)</script>"),
		"gif":       []byte("GIF89a\x01\x00\x01\x00\x00\x00\x00;"),
		"truncated": photo(t, 20, 20)[:40],
	} {
		if _, err := imaging.Recode(data); !errors.Is(err, imaging.ErrUnsupportedImage) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestRecode_RSEC70_DimensionBoundary(t *testing.T) {
	if _, err := imaging.Recode(photo(t, 4096, 1)); err != nil {
		t.Fatalf("4096 px: %v", err)
	}
	if _, err := imaging.Recode(photo(t, 4097, 1)); !errors.Is(err, imaging.ErrImageTooLarge) {
		t.Fatalf("4097 px: %v", err)
	}
	if _, err := imaging.Recode(photo(t, 1, 4097)); !errors.Is(err, imaging.ErrImageTooLarge) {
		t.Fatalf("4097 px tall: %v", err)
	}
}
