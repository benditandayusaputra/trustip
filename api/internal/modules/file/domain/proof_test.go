package domain

import (
	"bytes"
	"compress/zlib"
	"errors"
	"testing"
)

func pdfWith(body string) []byte {
	return []byte("%PDF-1.4\n" + body + "\ntrailer << /Root 1 0 R >>\n%%EOF\n")
}

func compressedStream(t *testing.T, content string) string {
	t.Helper()
	var packed bytes.Buffer
	writer := zlib.NewWriter(&packed)
	if _, err := writer.Write([]byte(content)); err != nil {
		t.Fatalf("compress: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	return "2 0 obj << /Type /ObjStm /Filter /FlateDecode >>\nstream\n" + packed.String() + "\nendstream\nendobj"
}

func TestCheckProofPDF_PlainReceiptIsKeptAsIs(t *testing.T) {
	receipt := pdfWith("1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj")

	document, err := CheckProofPDF(receipt)

	if err != nil || !bytes.Equal(document.Data, receipt) || document.ContentType != "application/pdf" ||
		document.Extension != "pdf" || document.Width != 0 || document.Height != 0 {
		t.Fatalf("document = %+v, err = %v", document, err)
	}
}

func TestCheckProofPDF_RSEC70_ActiveOrHiddenContentIsRejected(t *testing.T) {
	cases := map[string][]byte{
		"not a pdf":       []byte("GIF89a not a receipt"),
		"truncated":       []byte("%PDF-1.4\n1 0 obj << /Type /Catalog >> endobj\n"),
		"script":          pdfWith("1 0 obj << /OpenAction << /S /JavaScript /JS (app.alert(1)) >> >> endobj"),
		"escaped name":    pdfWith("1 0 obj << /OpenAction << /S /J#61vaScript /JS (x) >> >> endobj"),
		"launch":          pdfWith("1 0 obj << /OpenAction << /S /Launch /F (cmd.exe) >> >> endobj"),
		"attachment":      pdfWith("1 0 obj << /Names << /EmbeddedFiles 3 0 R >> >> endobj"),
		"locked":          pdfWith("1 0 obj << /Type /Catalog >> endobj\ntrailer << /Encrypt 4 0 R >>"),
		"compressed code": pdfWith(compressedStream(t, "<< /S /JavaScript /JS (app.alert(1)) >>")),
	}
	for name, data := range cases {
		if _, err := CheckProofPDF(data); !errors.Is(err, ErrUnsafeDocument) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestCheckProofPDF_CompressedPageContentIsAllowed(t *testing.T) {
	receipt := pdfWith(compressedStream(t, "BT /F1 12 Tf (Transfer Rp691.187 berhasil) Tj ET"))

	if _, err := CheckProofPDF(receipt); err != nil {
		t.Fatalf("err = %v", err)
	}
}
