package project

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectDomainScaffoldIsDiscoverableAndPure(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read project package directory: %v", err)
	}

	goFiles := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		goFiles[entry.Name()] = true
	}

	for _, name := range []string{
		"project.go",
		"projecttable.go",
		"projecttablerelation.go",
		"relationvaluesource.go",
		"validation.go",
	} {
		if !goFiles[name] {
			t.Fatalf("missing scaffold carrying file %s", name)
		}
	}
}

func TestProjectDomainScaffoldHasNoPrematurePublicContract(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob project package files: %v", err)
	}

	fset := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		parsed, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		if parsed.Name.Name != "project" {
			t.Fatalf("%s package name = %q, want project", file, parsed.Name.Name)
		}

		for _, decl := range parsed.Decls {
			switch typed := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range typed.Specs {
					switch typedSpec := spec.(type) {
					case *ast.TypeSpec:
						if typedSpec.Name.IsExported() {
							t.Fatalf("%s exports %s before model task boundaries", file, typedSpec.Name.Name)
						}
					case *ast.ValueSpec:
						for _, name := range typedSpec.Names {
							if name.IsExported() {
								t.Fatalf("%s exports %s before enum or validation task boundaries", file, name.Name)
							}
						}
					}
				}
			case *ast.FuncDecl:
				if typed.Name.IsExported() {
					t.Fatalf("%s exports %s before validation task boundaries", file, typed.Name.Name)
				}
			}
		}
	}
}

func TestProjectDomainScaffoldAvoidsOutOfScopeDependencies(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob project package files: %v", err)
	}

	forbidden := []string{
		"wails",
		"vue",
		"store",
		"service",
		"engine",
		"generator",
		"database/sql",
		"modernc.org/sqlite",
		"internal/dbx",
	}

	fset := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		parsed, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse imports for %s: %v", file, err)
		}

		for _, imported := range parsed.Imports {
			path := strings.Trim(imported.Path.Value, "\"")
			for _, blocked := range forbidden {
				if strings.Contains(path, blocked) {
					t.Fatalf("%s imports out-of-scope dependency %q", file, path)
				}
			}
		}
	}
}
