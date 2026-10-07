package main

import (
	"regexp"
	"strings"
)

var sqlDirective = regexp.MustCompile(
	`^-- (name: \w+ :\w+|\+goose (Up|Down|StatementBegin|StatementEnd|NO TRANSACTION))\s*$`,
)

func scanSQLComments(source []byte) []int {
	text := string(source)
	var lines []int
	for index := 0; index < len(text); index++ {
		next, isComment := skipSQLToken(text, index)
		if isComment {
			lines = append(lines, lineAt(text, index))
		}
		index = next
	}
	return lines
}

func skipSQLToken(text string, index int) (int, bool) {
	rest := text[index:]
	switch {
	case strings.HasPrefix(rest, "--"):
		end := lineEnd(text, index)
		return end, !sqlDirective.MatchString(text[index:end])
	case strings.HasPrefix(rest, "/*"):
		return blockEnd(text, index, "*/"), true
	case rest[0] == '\'' || rest[0] == '"':
		return sqlQuoteEnd(text, index), false
	case rest[0] == '$':
		return dollarQuoteEnd(text, index), false
	}
	return index, false
}

var dollarTag = regexp.MustCompile(`^\$[A-Za-z_]*\$`)

func dollarQuoteEnd(text string, index int) int {
	tag := dollarTag.FindString(text[index:])
	if tag == "" {
		return index
	}
	closing := strings.Index(text[index+len(tag):], tag)
	if closing < 0 {
		return len(text) - 1
	}
	return index + len(tag) + closing + len(tag) - 1
}

func sqlQuoteEnd(text string, index int) int {
	closing := strings.IndexByte(text[index+1:], text[index])
	if closing < 0 {
		return len(text) - 1
	}
	return index + 1 + closing
}
