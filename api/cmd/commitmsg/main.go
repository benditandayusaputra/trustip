package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maxSubjectLength  = 72
	variationSelector = "️"
)

var (
	errEmpty            = errors.New("commit message is empty")
	errSubjectShape     = errors.New("subject must be '<gitmoji> <Kalimat perintah>'")
	errSubjectLength    = errors.New("subject is longer than 72 characters")
	errSubjectPeriod    = errors.New("subject must not end with a period")
	errSubjectNumbering = errors.New("subject must describe the work, not a stage or feature number")
	errMissingBlankLine = errors.New("body must be separated from the subject by one blank line")
	errForeignTrailer   = errors.New("only the Rujukan trailer is allowed")
	errReferenceShape   = errors.New("the Rujukan trailer must list IDs separated by ', '")
	errForbiddenContent = errors.New("links, robot emoji, and generator notes are not allowed")
)

var (
	trailerLine     = regexp.MustCompile(`^[A-Za-z][A-Za-z-]*:\s`)
	referenceLine   = regexp.MustCompile(`^Rujukan: [A-Z0-9][A-Za-z0-9.-]*(, [A-Z0-9][A-Za-z0-9.-]*)*$`)
	numberedSubject = regexp.MustCompile(`(?i)\b(tahap|fitur)\s*\d`)
	forbiddenText   = regexp.MustCompile(`(?i)(https?://|generated with|\x{1F916})`)
)

func main() {
	message, err := io.ReadAll(os.Stdin)
	if err == nil {
		err = validateMessage(string(message))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "commit-msg:", err)
		os.Exit(1)
	}
}

func validateMessage(message string) error {
	lines := meaningfulLines(message)
	if len(lines) == 0 {
		return errEmpty
	}
	if err := validateSubject(lines[0]); err != nil {
		return err
	}
	if len(lines) > 1 && lines[1] != "" {
		return errMissingBlankLine
	}
	return validateBody(lines[1:])
}

func meaningfulLines(message string) []string {
	var lines []string
	for line := range strings.SplitSeq(message, "\n") {
		if !strings.HasPrefix(line, "#") {
			lines = append(lines, strings.TrimRight(line, " \t\r"))
		}
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func validateSubject(subject string) error {
	if utf8.RuneCountInString(subject) > maxSubjectLength {
		return errSubjectLength
	}
	if strings.HasSuffix(subject, ".") {
		return errSubjectPeriod
	}
	if numberedSubject.MatchString(subject) || forbiddenText.MatchString(subject) {
		return errSubjectNumbering
	}
	return validateSubjectShape(subject)
}

func validateSubjectShape(subject string) error {
	emoji, sentence, found := strings.Cut(strings.ReplaceAll(subject, variationSelector, ""), " ")
	if !found || !isGitmoji(emoji) {
		return errSubjectShape
	}
	first, _ := utf8.DecodeRuneInString(sentence)
	if !unicode.IsUpper(first) {
		return errSubjectShape
	}
	return nil
}

func validateBody(lines []string) error {
	for _, line := range lines {
		if forbiddenText.MatchString(line) {
			return errForbiddenContent
		}
		if err := validateTrailer(line); err != nil {
			return err
		}
	}
	return nil
}

func validateTrailer(line string) error {
	if !trailerLine.MatchString(line) {
		return nil
	}
	if !strings.HasPrefix(line, "Rujukan:") {
		return errForeignTrailer
	}
	if !referenceLine.MatchString(line) {
		return errReferenceShape
	}
	return nil
}

func isGitmoji(emoji string) bool {
	switch emoji {
	case "🎉", "✨", "🐛", "♻", "✅", "🔒", "🛂", "🦺", "⚡", "💄", "📱", "♿", "🚸", "💬",
		"🗃", "🌱", "🔊", "🥅", "🏗", "🧱", "🔧", "👷", "🚀", "➕", "➖", "📝", "🔥", "🚚":
		return true
	}
	return false
}
