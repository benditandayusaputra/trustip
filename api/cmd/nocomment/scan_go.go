package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func scanGoComments(source []byte) ([]int, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "", source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse go source: %w", err)
	}
	lines := []int{}
	if ast.IsGenerated(file) {
		return lines, nil
	}
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if !isGoDirective(comment.Text) {
				lines = append(lines, fileSet.Position(comment.Pos()).Line)
			}
		}
	}
	return lines, nil
}

func isGoDirective(text string) bool {
	for _, directive := range []string{"//go:build", "//go:embed", "//go:generate"} {
		if strings.HasPrefix(text, directive) {
			return true
		}
	}
	return false
}
