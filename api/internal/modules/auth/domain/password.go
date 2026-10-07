package domain

import (
	_ "embed"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

const (
	MinPasswordLength      = 10
	MaxPasswordLength      = 128
	minIdentityTokenLength = 3
	passwordField          = "password"
)

//go:embed common_passwords.txt
var commonPasswordList string

type PasswordCandidate struct {
	Password string
	FullName string
	Email    string
}

type PasswordPolicy struct {
	common map[string]struct{}
}

func NewPasswordPolicy() PasswordPolicy {
	lines := strings.Fields(commonPasswordList)
	common := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		common[line] = struct{}{}
	}
	return PasswordPolicy{common: common}
}

func (p PasswordPolicy) FindProblem(candidate PasswordCandidate) (apperror.FieldError, bool) {
	length := utf8.RuneCountInString(candidate.Password)
	lowered := strings.ToLower(candidate.Password)
	switch {
	case length < MinPasswordLength:
		return lengthProblem(apperror.FieldTooShort, "min", MinPasswordLength), true
	case length > MaxPasswordLength:
		return lengthProblem(apperror.FieldTooLong, "max", MaxPasswordLength), true
	case p.isCommon(lowered):
		return apperror.InvalidField(passwordField, apperror.FieldTooCommon), true
	case containsAny(lowered, identityTokens(candidate)):
		return apperror.InvalidField(passwordField, apperror.FieldContainsIdentity), true
	}
	return apperror.FieldError{}, false
}

func lengthProblem(code apperror.FieldCode, param string, limit int) apperror.FieldError {
	return apperror.InvalidField(passwordField, code, apperror.Param{Name: param, Value: strconv.Itoa(limit)})
}

func (p PasswordPolicy) isCommon(lowered string) bool {
	_, found := p.common[lowered]
	return found
}

func identityTokens(candidate PasswordCandidate) []string {
	email := strings.ToLower(candidate.Email)
	localPart, _, _ := strings.Cut(email, "@")
	name := strings.ToLower(candidate.FullName)
	tokens := []string{email, localPart, strings.Join(strings.Fields(name), "")}
	tokens = append(tokens, words(name)...)
	tokens = append(tokens, words(localPart)...)
	return tokens
}

func words(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func containsAny(text string, tokens []string) bool {
	for _, token := range tokens {
		if utf8.RuneCountInString(token) >= minIdentityTokenLength && strings.Contains(text, token) {
			return true
		}
	}
	return false
}
