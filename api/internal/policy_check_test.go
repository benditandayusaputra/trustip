package internal_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

const useCaseMethod = "Execute"

func sessionModules() map[string]bool {
	return map[string]bool{"auth": true}
}

func TestArchitecture_RSEC36_EveryUseCaseRequiresAPolicyFirst(t *testing.T) {
	for _, file := range loadSourceFiles(t) {
		if file.layer.kind != "service" || sessionModules()[file.layer.module] {
			continue
		}
		for _, decl := range file.decls {
			if isUseCase(decl) && !startsWithPolicyRequire(decl) {
				t.Errorf("%s: %s must call policy.Require in its first statement", file.path, receiverName(decl))
			}
		}
	}
}

func TestStartsWithPolicyRequire(t *testing.T) {
	cases := map[string]bool{
		`if err := policy.Require(signedIn, policy.EditProfile); err != nil { return err }`: true,
		`policy.Require(signedIn, policy.EditProfile)`:                                      false,
		`if err := other.Require(signedIn, policy.EditProfile); err != nil { return err }`:  false,
		`if err := load(); err != nil { return err }`:                                       false,
		``: false,
	}
	for body, want := range cases {
		decl := parseUseCase(t, body)
		if got := startsWithPolicyRequire(decl); got != want {
			t.Errorf("startsWithPolicyRequire(%q) = %v, want %v", body, got, want)
		}
	}
}

func parseUseCase(t *testing.T, body string) *ast.FuncDecl {
	t.Helper()
	source := "package service\nfunc (uc *UseCase) Execute() error {\n" + body + "\nreturn nil\n}\n"
	parsed, err := parser.ParseFile(token.NewFileSet(), "", source, 0)
	if err != nil {
		t.Fatalf("parse %q: %v", body, err)
	}
	return funcDecls(parsed)[0]
}

func isUseCase(decl *ast.FuncDecl) bool {
	return decl.Recv != nil && decl.Name.Name == useCaseMethod
}

func receiverName(decl *ast.FuncDecl) string {
	receiver := decl.Recv.List[0].Type
	if star, ok := receiver.(*ast.StarExpr); ok {
		receiver = star.X
	}
	if ident, ok := receiver.(*ast.Ident); ok {
		return ident.Name + "." + useCaseMethod
	}
	return useCaseMethod
}

func startsWithPolicyRequire(decl *ast.FuncDecl) bool {
	if decl.Body == nil || len(decl.Body.List) == 0 {
		return false
	}
	guard, ok := decl.Body.List[0].(*ast.IfStmt)
	if !ok {
		return false
	}
	assign, ok := guard.Init.(*ast.AssignStmt)
	if !ok || len(assign.Rhs) != 1 {
		return false
	}
	return isPolicyRequireCall(assign.Rhs[0])
}

func isPolicyRequireCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := selector.X.(*ast.Ident)
	return ok && pkg.Name == "policy" && selector.Sel.Name == "Require"
}
