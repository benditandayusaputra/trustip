package httpserver

import "github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"

type TextInput struct {
	Name  string
	Value string
	clean bool
}

func LineInput(name, raw string) TextInput {
	value, clean := NormalizeLine(raw)
	return TextInput{Name: name, Value: value, clean: clean}
}

func RejectUncleanInputs(inputs ...TextInput) error {
	var problems []apperror.FieldError
	for _, input := range inputs {
		if !input.clean {
			problems = append(problems, apperror.InvalidField(input.Name, apperror.FieldInvalidCharacters))
		}
	}
	if len(problems) > 0 {
		return apperror.Validation(problems...)
	}
	return nil
}

func MultilineInput(name, raw string) TextInput {
	value, clean := NormalizeMultiline(raw)
	return TextInput{Name: name, Value: value, clean: clean}
}
