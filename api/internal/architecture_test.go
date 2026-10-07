package internal_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	modulePath      = "github.com/benditandayusaputra/trustip/api"
	internalPath    = modulePath + "/internal"
	kernelPath      = internalPath + "/kernel"
	modulesPath     = internalPath + "/modules"
	sqlcPath        = internalPath + "/db/sqlc"
	databasePath    = internalPath + "/platform/database"
	httpServerPath  = internalPath + "/platform/httpserver"
	fiberPath       = "github.com/gofiber/fiber"
	pgxPath         = "github.com/jackc/pgx"
	maxParameters   = 3
	maxResultValues = 2
)

type layer struct {
	kind   string
	module string
}

type sourceFile struct {
	path    string
	ownPath string
	layer   layer
	imports []string
	decls   []*ast.FuncDecl
}

func TestArchitecture_ImportsFollowLayerRules(t *testing.T) {
	for _, file := range loadSourceFiles(t) {
		for _, imported := range file.imports {
			if imported != file.ownPath && !isImportAllowed(file.layer, imported) {
				t.Errorf("%s (%s) must not import %s", file.path, file.layer.kind, imported)
			}
		}
	}
}

func TestArchitecture_FunctionsRespectParameterLimit(t *testing.T) {
	for _, file := range loadSourceFiles(t) {
		for _, decl := range file.decls {
			if count := countParametersWithoutContext(decl.Type); count > maxParameters {
				t.Errorf("%s: %s has %d parameters besides ctx, max %d",
					file.path, decl.Name.Name, count, maxParameters)
			}
		}
	}
}

func TestArchitecture_FunctionsRespectResultLimit(t *testing.T) {
	for _, file := range loadSourceFiles(t) {
		for _, decl := range file.decls {
			if count := countResultsWithoutError(decl.Type); count > maxResultValues {
				t.Errorf("%s: %s returns %d values besides error, max %d",
					file.path, decl.Name.Name, count, maxResultValues)
			}
		}
	}
}

func TestIsImportAllowed(t *testing.T) {
	cases := []struct {
		name     string
		layer    layer
		imported string
		want     bool
	}{
		{"domain imports stdlib", layer{"domain", "drop"}, "time", true},
		{"domain imports kernel", layer{"domain", "drop"}, kernelPath + "/money", true},
		{"domain imports fiber", layer{"domain", "drop"}, fiberPath + "/v3", false},
		{"domain imports other module", layer{"domain", "drop"}, modulesPath + "/stock/domain", false},
		{"kernel imports pgx", layer{"kernel", ""}, pgxPath + "/v5", false},
		{"service imports own domain", layer{"service", "drop"}, modulesPath + "/drop/domain", true},
		{"service imports other domain", layer{"service", "drop"}, modulesPath + "/stock/domain", false},
		{"service imports repository", layer{"service", "drop"}, modulesPath + "/drop/repository", false},
		{"service imports platform", layer{"service", "drop"}, databasePath, false},
		{"repository imports sqlc", layer{"repository", "drop"}, sqlcPath, true},
		{"repository imports pgx", layer{"repository", "drop"}, pgxPath + "/v5", true},
		{"repository imports fiber", layer{"repository", "drop"}, fiberPath + "/v3", false},
		{"repository imports service", layer{"repository", "drop"}, modulesPath + "/drop/service", false},
		{"handler imports own service", layer{"handler", "drop"}, modulesPath + "/drop/service", true},
		{"handler imports httpserver", layer{"handler", "drop"}, httpServerPath, true},
		{"handler imports fiber", layer{"handler", "drop"}, fiberPath + "/v3", true},
		{"handler imports sqlc", layer{"handler", "drop"}, sqlcPath, false},
		{"handler imports own domain", layer{"handler", "drop"}, modulesPath + "/drop/domain", true},
		{"handler imports other domain", layer{"handler", "drop"}, modulesPath + "/stock/domain", false},
		{"platform imports pgx", layer{"platform", "database"}, pgxPath + "/v5", true},
		{"platform imports modules", layer{"platform", "database"}, modulesPath + "/drop/domain", false},
		{"unknown layer imports anything", layer{"other", ""}, fiberPath + "/v3", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isImportAllowed(tc.layer, tc.imported); got != tc.want {
				t.Fatalf("isImportAllowed = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLayerOf(t *testing.T) {
	cases := map[string]layer{
		"kernel/money":                {"kernel", ""},
		"modules/drop/domain":         {"domain", "drop"},
		"modules/drop/service":        {"service", "drop"},
		"modules/drop/service/policy": {"service", "drop"},
		"modules/drop/repository":     {"repository", "drop"},
		"modules/drop/handler":        {"handler", "drop"},
		"platform/database":           {"platform", "database"},
		"db/sqlc":                     {"other", ""},
		".":                           {"other", ""},
	}
	for dir, want := range cases {
		if got := layerOf(dir); got != want {
			t.Errorf("layerOf(%q) = %v, want %v", dir, got, want)
		}
	}
}

func layerOf(dir string) layer {
	parts := strings.Split(filepath.ToSlash(dir), "/")
	const moduleLayerDepth = 3
	switch {
	case parts[0] == "kernel":
		return layer{"kernel", ""}
	case parts[0] == "platform" && len(parts) > 1:
		return layer{"platform", parts[1]}
	case parts[0] == "modules" && len(parts) >= moduleLayerDepth:
		return layer{parts[2], parts[1]}
	}
	return layer{"other", ""}
}

func isImportAllowed(owner layer, imported string) bool {
	if isStandardLibrary(imported) || strings.HasPrefix(imported, kernelPath) {
		return true
	}
	ownModule := modulesPath + "/" + owner.module + "/"
	switch owner.kind {
	case "kernel", "domain":
		return false
	case "service":
		return strings.HasPrefix(imported, ownModule+"domain")
	case "repository":
		return hasAnyPrefix(imported, ownModule+"domain", sqlcPath, databasePath, pgxPath)
	case "handler":
		return hasAnyPrefix(imported, ownModule+"service", ownModule+"domain", httpServerPath, fiberPath)
	case "platform":
		return !strings.HasPrefix(imported, modulesPath)
	}
	return true
}

func isStandardLibrary(imported string) bool {
	firstElement, _, _ := strings.Cut(imported, "/")
	return !strings.Contains(firstElement, ".")
}

func hasAnyPrefix(value string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func countParametersWithoutContext(fn *ast.FuncType) int {
	count := 0
	for _, field := range fn.Params.List {
		if isContextType(field.Type) {
			continue
		}
		count += max(len(field.Names), 1)
	}
	return count
}

func isContextType(expr ast.Expr) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := selector.X.(*ast.Ident)
	return ok && pkg.Name == "context" && selector.Sel.Name == "Context"
}

func countResultsWithoutError(fn *ast.FuncType) int {
	if fn.Results == nil {
		return 0
	}
	count := 0
	for _, field := range fn.Results.List {
		if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "error" {
			continue
		}
		count += max(len(field.Names), 1)
	}
	return count
}

func loadSourceFiles(t *testing.T) []sourceFile {
	t.Helper()
	var files []sourceFile
	err := filepath.WalkDir("..", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		file, ok := parseSourceFile(t, path)
		if ok {
			files = append(files, file)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk sources: %v", err)
	}
	return files
}

func parseSourceFile(t *testing.T, path string) (sourceFile, bool) {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	if ast.IsGenerated(parsed) {
		return sourceFile{}, false
	}
	return sourceFile{
		path:    path,
		ownPath: importPathOf(path),
		layer:   layerOf(relativeToInternal(path)),
		imports: importPaths(parsed),
		decls:   funcDecls(parsed),
	}, true
}

func relativeToInternal(path string) string {
	relative, err := filepath.Rel(filepath.Join("..", "internal"), filepath.Dir(path))
	if err != nil || strings.HasPrefix(relative, "..") {
		return "."
	}
	return relative
}

func importPathOf(path string) string {
	return modulePath + "/" + filepath.ToSlash(filepath.Dir(strings.TrimPrefix(path, "../")))
}

func importPaths(file *ast.File) []string {
	paths := make([]string, 0, len(file.Imports))
	for _, spec := range file.Imports {
		if path, err := strconv.Unquote(spec.Path.Value); err == nil {
			paths = append(paths, path)
		}
	}
	return paths
}

func funcDecls(file *ast.File) []*ast.FuncDecl {
	var decls []*ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			decls = append(decls, fn)
		}
	}
	return decls
}
