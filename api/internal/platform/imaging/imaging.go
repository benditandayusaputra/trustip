package imaging

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
)

const (
	MaxDimension = 4096
	jpegQuality  = 85
)

var (
	ErrUnsupportedImage = errors.New("content is not a JPEG or PNG image")
	ErrImageTooLarge    = errors.New("image is wider or taller than allowed")
)

type Recoded struct {
	Data        []byte
	ContentType string
	Extension   string
	Width       int
	Height      int
}

type format struct {
	contentType string
	extension   string
	encode      func(*bytes.Buffer, image.Image) error
}

func formats() map[string]format {
	return map[string]format{
		"jpeg": {contentType: "image/jpeg", extension: "jpg", encode: encodeJPEG},
		"png":  {contentType: "image/png", extension: "png", encode: encodePNG},
	}
}

func Recode(data []byte) (Recoded, error) {
	config, name, err := image.DecodeConfig(bytes.NewReader(data))
	chosen, known := formats()[name]
	if err != nil || !known {
		return Recoded{}, ErrUnsupportedImage
	}
	if config.Width > MaxDimension || config.Height > MaxDimension {
		return Recoded{}, ErrImageTooLarge
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return Recoded{}, ErrUnsupportedImage
	}
	var encoded bytes.Buffer
	if err := chosen.encode(&encoded, decoded); err != nil {
		return Recoded{}, fmt.Errorf("encode %s: %w", name, err)
	}
	return Recoded{
		Data: encoded.Bytes(), ContentType: chosen.contentType, Extension: chosen.extension,
		Width: config.Width, Height: config.Height,
	}, nil
}

func encodeJPEG(target *bytes.Buffer, decoded image.Image) error {
	if err := jpeg.Encode(target, decoded, &jpeg.Options{Quality: jpegQuality}); err != nil {
		return fmt.Errorf("write jpeg: %w", err)
	}
	return nil
}

func encodePNG(target *bytes.Buffer, decoded image.Image) error {
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(target, decoded); err != nil {
		return fmt.Errorf("write png: %w", err)
	}
	return nil
}
