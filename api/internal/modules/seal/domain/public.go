package domain

import (
	"bytes"
	"strings"
)

const codeCharacters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

func NormalizeCode(raw string) (string, bool) {
	compact := strings.Map(func(r rune) rune {
		if r == ' ' || r == '-' {
			return -1
		}
		return r
	}, strings.ToUpper(raw))
	if len(compact) != codeLength || strings.Trim(compact, codeCharacters) != "" {
		return "", false
	}
	groups := make([]string, 0, codeLength/codeGroupLength)
	for start := 0; start < codeLength; start += codeGroupLength {
		groups = append(groups, compact[start:start+codeGroupLength])
	}
	return strings.Join(groups, codeSeparator), true
}

type Parties struct {
	ConsignorName string
	StoreName     string
}

type PublicSeal struct {
	Seal              Seal
	Parties           Parties
	PreviousEntryHash []byte
}

func (p PublicSeal) IsLinked() bool {
	return bytes.Equal(p.Seal.Entry.PrevHash, p.PreviousEntryHash)
}

type PublicVerdict struct {
	Seal    Seal
	Parties Parties
	IsValid bool
}

func JudgePublic(found PublicSeal, isSigned func(Seal) bool) PublicVerdict {
	isValid := found.Seal.CheckEntry() == nil && isSigned(found.Seal) && found.IsLinked()
	return PublicVerdict{Seal: found.Seal, Parties: found.Parties, IsValid: isValid}
}

func IsIntactChain(chain []Seal, isSigned func(Seal) bool) bool {
	if VerifyChain(chain) != nil {
		return false
	}
	for _, seal := range chain {
		if !isSigned(seal) {
			return false
		}
	}
	return true
}

type PublicKey struct {
	KeyID string
	Bytes []byte
}
