package field

import (
	"slices"
	"strconv"
	"unicode/utf8"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

type Problem struct {
	field apperror.FieldError
	found bool
}

func Invalid(name string, code apperror.FieldCode, params ...apperror.Param) Problem {
	return Problem{field: apperror.InvalidField(name, code, params...), found: true}
}

func Collect(candidates ...Problem) []apperror.FieldError {
	var problems []apperror.FieldError
	for _, candidate := range candidates {
		if candidate.found {
			problems = append(problems, candidate.field)
		}
	}
	return problems
}

type Length struct {
	Min int
	Max int
}

func CheckLength(name, value string, allowed Length) Problem {
	length := utf8.RuneCountInString(value)
	if length < allowed.Min {
		return Invalid(name, apperror.FieldTooShort, Limit("min", allowed.Min))
	}
	if length > allowed.Max {
		return Invalid(name, apperror.FieldTooLong, Limit("max", allowed.Max))
	}
	return Problem{}
}

func CheckChoice(name, value string, choices []string) Problem {
	if value == "" || slices.Contains(choices, value) {
		return Problem{}
	}
	return Invalid(name, apperror.FieldUnknownChoice)
}

func Limit(name string, limit int) apperror.Param {
	return apperror.Param{Name: name, Value: strconv.Itoa(limit)}
}
