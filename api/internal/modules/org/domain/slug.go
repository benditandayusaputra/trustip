package domain

import (
	"strconv"
	"strings"
	"unicode"
)

const (
	MaxSlugLength          = 60
	fallbackSlug      Slug = "usaha"
	firstSlugSuffix        = 2
	slugWordSeparator      = "-"
	accentedLetters        = "àáâãäåèéêëìíîïòóôõöùúûüñçýÿ"
	plainLetters           = "aaaaaaeeeeiiiiooooouuuuncyy"
)

type Slug string

func SlugFrom(name string) Slug {
	var kept []string
	length := 0
	for _, word := range slugWords(name) {
		grown := length + len(word) + len(slugWordSeparator)*min(len(kept), 1)
		if grown > MaxSlugLength {
			break
		}
		kept = append(kept, word)
		length = grown
	}
	if len(kept) == 0 {
		return firstWordOrFallback(slugWords(name))
	}
	return Slug(strings.Join(kept, slugWordSeparator))
}

func firstWordOrFallback(words []string) Slug {
	if len(words) == 0 {
		return fallbackSlug
	}
	return Slug(words[0][:MaxSlugLength])
}

func slugWords(name string) []string {
	folded := strings.Map(foldToSlugRune, strings.ToLower(name))
	return strings.FieldsFunc(folded, func(r rune) bool { return r == ' ' })
}

func foldToSlugRune(r rune) rune {
	if r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
		return r
	}
	accented := []rune(accentedLetters)
	for i, letter := range accented {
		if letter == r {
			return []rune(plainLetters)[i]
		}
	}
	return ' '
}

func NextFreeSlug(base Slug, taken []Slug) Slug {
	isTaken := map[Slug]bool{}
	for _, slug := range taken {
		isTaken[slug] = true
	}
	candidate := base
	for suffix := firstSlugSuffix; isTaken[candidate]; suffix++ {
		candidate = base + Slug(slugWordSeparator+strconv.Itoa(suffix))
	}
	return candidate
}
