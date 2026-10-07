package canonical

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"unicode/utf16"
	"unicode/utf8"
)

const firstPrintable = 0x20

var ErrUnsupported = errors.New("canonical form holds only objects, arrays, strings, and integers")

func Encode(value any) ([]byte, error) {
	var out bytes.Buffer
	if err := write(&out, value); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func Digest(value any) ([]byte, error) {
	encoded, err := Encode(value)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(encoded)
	return sum[:], nil
}

func write(out *bytes.Buffer, value any) error {
	switch typed := value.(type) {
	case map[string]any:
		return writeObject(out, typed)
	case []any:
		return writeArray(out, typed)
	case string:
		return writeString(out, typed)
	case int:
		out.WriteString(strconv.Itoa(typed))
	case int64:
		out.WriteString(strconv.FormatInt(typed, 10))
	default:
		return fmt.Errorf("%w: %T", ErrUnsupported, value)
	}
	return nil
}

func writeObject(out *bytes.Buffer, object map[string]any) error {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, compareUTF16)
	out.WriteByte('{')
	for index, key := range keys {
		if index > 0 {
			out.WriteByte(',')
		}
		if err := writeString(out, key); err != nil {
			return err
		}
		out.WriteByte(':')
		if err := write(out, object[key]); err != nil {
			return err
		}
	}
	out.WriteByte('}')
	return nil
}

func compareUTF16(left, right string) int {
	return slices.Compare(utf16.Encode([]rune(left)), utf16.Encode([]rune(right)))
}

func writeArray(out *bytes.Buffer, array []any) error {
	out.WriteByte('[')
	for index, element := range array {
		if index > 0 {
			out.WriteByte(',')
		}
		if err := write(out, element); err != nil {
			return err
		}
	}
	out.WriteByte(']')
	return nil
}

func shortEscapes() map[rune]string {
	return map[rune]string{
		'"': `\"`, '\\': `\\`, '\b': `\b`, '\f': `\f`, '\n': `\n`, '\r': `\r`, '\t': `\t`,
	}
}

func writeString(out *bytes.Buffer, text string) error {
	if !utf8.ValidString(text) {
		return fmt.Errorf("%w: invalid utf-8", ErrUnsupported)
	}
	escapes := shortEscapes()
	out.WriteByte('"')
	for _, char := range text {
		if escaped, ok := escapes[char]; ok {
			out.WriteString(escaped)
			continue
		}
		if char < firstPrintable {
			fmt.Fprintf(out, `\u%04x`, char)
			continue
		}
		out.WriteRune(char)
	}
	out.WriteByte('"')
	return nil
}
