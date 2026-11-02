package domain

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"errors"
	"io"
	"regexp"
)

const (
	pdfHeader           = "%PDF-"
	pdfEnd              = "%%EOF"
	pdfContentType      = "application/pdf"
	pdfExtension        = "pdf"
	pdfTrailerWindow    = 1024
	maxInflatedPDFBytes = 8 << 20
)

var (
	ErrUnsafeDocument = errors.New("upload is not a plain pdf")

	pdfStream      = regexp.MustCompile(`(?s)(?:^|[^d])stream\r?\n(.*?)endstream`)
	pdfHexEscape   = regexp.MustCompile(`#[0-9A-Fa-f]{2}`)
	pdfActiveNames = regexp.MustCompile(
		`/(JavaScript|JS|Launch|EmbeddedFiles?|RichMedia|XFA|SubmitForm|ImportData|GoToE|GoToR|Encrypt)` +
			`([\s/<>\[\]()%{}]|$)`,
	)
)

func IsPDF(data []byte) bool {
	return bytes.HasPrefix(data, []byte(pdfHeader))
}

func CheckProofPDF(data []byte) (Image, error) {
	trailer := data[max(0, len(data)-pdfTrailerWindow):]
	if !IsPDF(data) || !bytes.Contains(trailer, []byte(pdfEnd)) || hasActiveContent(data) {
		return Image{}, ErrUnsafeDocument
	}
	inflated, err := inflateStreams(data)
	if err != nil || hasActiveContent(inflated) {
		return Image{}, ErrUnsafeDocument
	}
	return Image{Data: data, ContentType: pdfContentType, Extension: pdfExtension}, nil
}

func hasActiveContent(data []byte) bool {
	plain := pdfHexEscape.ReplaceAllFunc(data, func(escape []byte) []byte {
		decoded, err := hex.DecodeString(string(escape[1:]))
		if err != nil {
			return escape
		}
		return decoded
	})
	return pdfActiveNames.Match(plain)
}

func inflateStreams(data []byte) ([]byte, error) {
	var inflated bytes.Buffer
	for _, match := range pdfStream.FindAllSubmatch(data, -1) {
		inflateInto(&inflated, match[1], int64(maxInflatedPDFBytes-inflated.Len()+1))
		if inflated.Len() > maxInflatedPDFBytes {
			return nil, ErrUnsafeDocument
		}
	}
	return inflated.Bytes(), nil
}

func inflateInto(target *bytes.Buffer, packed []byte, budget int64) {
	reader, err := zlib.NewReader(bytes.NewReader(packed))
	if err != nil {
		return
	}
	_, _ = io.Copy(target, io.LimitReader(reader, budget))
}
