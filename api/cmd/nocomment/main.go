package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"
)

var errCommentsFound = errors.New("comments found")

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	root, err := repositoryRoot(ctx)
	if err != nil {
		return err
	}
	paths, err := trackedFiles(ctx, root)
	if err != nil {
		return err
	}
	findings, err := scanFiles(os.DirFS(root), paths)
	if err != nil {
		return err
	}
	return report(out, findings)
}

func repositoryRoot(ctx context.Context) (string, error) {
	output, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("find repository root: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func trackedFiles(ctx context.Context, root string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list repository files: %w", err)
	}
	var paths []string
	for _, path := range bytes.Split(output, []byte{0}) {
		if len(path) > 0 {
			paths = append(paths, string(path))
		}
	}
	return paths, nil
}

func scanFiles(repository fs.FS, paths []string) ([]Finding, error) {
	var findings []Finding
	for _, path := range paths {
		fileFindings, err := scanFile(repository, path)
		if err != nil {
			return nil, err
		}
		findings = append(findings, fileFindings...)
	}
	return findings, nil
}

func scanFile(repository fs.FS, path string) ([]Finding, error) {
	scan, err := scannerFor(path)
	if err != nil {
		return nil, err
	}
	source, err := fs.ReadFile(repository, path)
	if errors.Is(err, fs.ErrNotExist) {
		return []Finding{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	lines, err := scan(source)
	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}
	return withPath(path, lines), nil
}

func withPath(path string, lines []int) []Finding {
	findings := make([]Finding, 0, len(lines))
	for _, line := range lines {
		findings = append(findings, Finding{Path: path, Line: line})
	}
	return findings
}

func report(out io.Writer, findings []Finding) error {
	for _, finding := range findings {
		if _, err := fmt.Fprintf(out, "%s:%d: comment is not allowed\n", finding.Path, finding.Line); err != nil {
			return fmt.Errorf("write report: %w", err)
		}
	}
	if len(findings) > 0 {
		return fmt.Errorf("%d %w", len(findings), errCommentsFound)
	}
	return nil
}

type Finding struct {
	Path string
	Line int
}
