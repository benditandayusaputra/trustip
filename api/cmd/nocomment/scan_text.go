package main

import (
	"regexp"
	"strings"
)

func scanCSSComments(source []byte) []int {
	text := string(source)
	var lines []int
	for index := 0; index < len(text); index++ {
		switch {
		case strings.HasPrefix(text[index:], "/*"):
			lines = append(lines, lineAt(text, index))
			index = blockEnd(text, index, "*/")
		case text[index] == '\'' || text[index] == '"':
			index = quoteEnd(text, index)
		}
	}
	return lines
}

func scanMarkupComments(source []byte) []int {
	text := string(source)
	var lines []int
	for offset := 0; ; {
		found := strings.Index(text[offset:], "<!--")
		if found < 0 {
			return lines
		}
		lines = append(lines, lineAt(text, offset+found))
		offset = blockEnd(text, offset+found, "-->") + 1
	}
}

var svelteStyleBlock = regexp.MustCompile(`(?s)<style[^>]*>(.*?)</style>`)

func scanSvelteStyleComments(source []byte) []int {
	text := string(source)
	var lines []int
	for _, match := range svelteStyleBlock.FindAllStringSubmatchIndex(text, -1) {
		const contentStartGroup = 2
		start := match[contentStartGroup]
		offsetLine := lineAt(text, start) - 1
		for _, line := range scanCSSComments([]byte(text[start:match[contentStartGroup+1]])) {
			lines = append(lines, line+offsetLine)
		}
	}
	return lines
}

func scanHashComments(source []byte) []int {
	var lines []int
	for number, line := range strings.Split(string(source), "\n") {
		lineNumber := number + 1
		if lineNumber == 1 && strings.HasPrefix(line, "#!") {
			continue
		}
		if hasHashComment(line) {
			lines = append(lines, lineNumber)
		}
	}
	return lines
}

func hasHashComment(line string) bool {
	var quote rune
	previous := ' '
	for _, char := range line {
		switch {
		case quote != 0 && char == quote:
			quote = 0
		case quote == 0 && (char == '\'' || char == '"'):
			quote = char
		case quote == 0 && char == '#' && isSpace(previous):
			return true
		}
		previous = char
	}
	return false
}

func isSpace(char rune) bool {
	return char == ' ' || char == '\t'
}

func lineAt(text string, index int) int {
	return strings.Count(text[:index], "\n") + 1
}

func lineEnd(text string, index int) int {
	end := strings.IndexByte(text[index:], '\n')
	if end < 0 {
		return len(text)
	}
	return index + end
}

func blockEnd(text string, index int, closing string) int {
	end := strings.Index(text[index+len(closing):], closing)
	if end < 0 {
		return len(text) - 1
	}
	return index + len(closing) + end + len(closing) - 1
}

func quoteEnd(text string, index int) int {
	quote := text[index]
	for cursor := index + 1; cursor < len(text); cursor++ {
		if text[cursor] == '\\' {
			cursor++
			continue
		}
		if text[cursor] == quote {
			return cursor
		}
	}
	return len(text) - 1
}
