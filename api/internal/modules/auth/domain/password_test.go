package domain

import (
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

func candidate(password string) PasswordCandidate {
	return candidateOf(password, "Sari Wulandari", "sari.w@contoh.id")
}

func candidateOf(password, fullName, email string) PasswordCandidate {
	return PasswordCandidate{Password: password, FullName: fullName, Email: email}
}

func TestPasswordPolicy_RSEC01_LengthBoundaries(t *testing.T) {
	policy := NewPasswordPolicy()
	cases := []struct {
		name     string
		password string
		want     apperror.FieldCode
	}{
		{"R-SEC-01 nine characters is too short", "Kp7#mQz2v", apperror.FieldTooShort},
		{"R-SEC-01 ten characters is accepted", "Kp7#mQz2vX", ""},
		{"R-SEC-01 counts characters not bytes", "ĶƥŦ#ɱǬż2v", apperror.FieldTooShort},
		{"R-SEC-01 128 characters is accepted", strings.Repeat("Kp7#mQz2", 16), ""},
		{"R-SEC-01 129 characters is too long", strings.Repeat("Kp7#mQz2", 16) + "x", apperror.FieldTooLong},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			problem, found := policy.FindProblem(candidate(tc.password))

			if apperror.FieldCode(problem.Code) != tc.want || found != (tc.want != "") {
				t.Fatalf("problem = %+v, found = %v, want %s", problem, found, tc.want)
			}
		})
	}
}

func TestPasswordPolicy_RSEC01_TooShortNamesTheMinimum(t *testing.T) {
	problem, _ := NewPasswordPolicy().FindProblem(candidate("pendek"))

	if problem.Field != "password" || problem.Message != "Minimal 10 karakter." {
		t.Fatalf("problem = %+v", problem)
	}
}

func TestPasswordPolicy_RSEC01_CommonPasswordsAreRejected(t *testing.T) {
	policy := NewPasswordPolicy()
	for _, common := range []string{"1234567890", "qwertyuiop", "Password123", "iloveyou12"} {
		t.Run("R-SEC-01 "+common, func(t *testing.T) {
			problem, found := policy.FindProblem(candidate(common))

			if !found || apperror.FieldCode(problem.Code) != apperror.FieldTooCommon {
				t.Fatalf("problem = %+v, found = %v", problem, found)
			}
		})
	}
}

func TestPasswordPolicy_RSEC01_PasswordsWithIdentityAreRejected(t *testing.T) {
	policy := NewPasswordPolicy()
	cases := map[string]string{
		"R-SEC-01 first name":       "Kp7#SARIq2v",
		"R-SEC-01 last name":        "wulandari#2026",
		"R-SEC-01 joined full name": "sariwulandari!",
		"R-SEC-01 email local part": "x9Sari.W#77q",
		"R-SEC-01 whole email":      "sari.w@contoh.id",
	}
	for name, password := range cases {
		t.Run(name, func(t *testing.T) {
			problem, found := policy.FindProblem(candidate(password))

			if !found || apperror.FieldCode(problem.Code) != apperror.FieldContainsIdentity {
				t.Fatalf("problem = %+v, found = %v", problem, found)
			}
		})
	}
}

func TestPasswordPolicy_RSEC01_ShortNamePartsAreIgnored(t *testing.T) {
	policy := NewPasswordPolicy()
	short := candidateOf("jo#Kp7mQz2vX", "Jo Ra", "jo@ra.id")

	if problem, found := policy.FindProblem(short); found {
		t.Fatalf("problem = %+v", problem)
	}
}
