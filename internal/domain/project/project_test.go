package project

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestProjectAggregateRootExposesStableContractFields(t *testing.T) {
	projectType := reflect.TypeOf(Project{})

	assertProjectJSONTags(t, projectType, map[string]string{
		"ID":           "id",
		"ConnectionID": "connectionId",
		"Name":         "name",
		"Description":  "description",
		"CreatedAt":    "createdAt",
		"UpdatedAt":    "updatedAt",
	})
	assertProjectStructJSONFieldSet(t, projectType, []string{"id", "connectionId", "name", "description", "createdAt", "updatedAt"})

	assertProjectFieldType(t, projectType, "ID", reflect.TypeOf(int64(0)))
	assertProjectFieldType(t, projectType, "ConnectionID", reflect.TypeOf(int64(0)))
	assertProjectFieldType(t, projectType, "Name", reflect.TypeOf(""))
	assertProjectFieldType(t, projectType, "Description", reflect.TypeOf(""))
	assertProjectFieldType(t, projectType, "CreatedAt", reflect.TypeOf(time.Time{}))
	assertProjectFieldType(t, projectType, "UpdatedAt", reflect.TypeOf(time.Time{}))
}

func TestProjectJSONRoundTripPreservesAggregateRootFields(t *testing.T) {
	createdAt := time.Date(2026, 6, 10, 9, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 6, 10, 10, 45, 0, 0, time.UTC)
	original := Project{
		ID:           101,
		ConnectionID: 202,
		Name:         "Reporting Demo",
		Description:  "Reusable generation setup for reporting demos.",
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal(Project) returned error: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("Unmarshal encoded Project into field map returned error: %v", err)
	}
	assertProjectJSONFieldsPresent(t, fields, "id", "connectionId", "name", "description", "createdAt", "updatedAt")
	assertProjectJSONFieldsAbsent(t, fields, "connection_id", "created_at", "updated_at", "projectTables", "tables", "relations", "executionStatus", "generatorConfig")

	var decoded Project
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(Project) returned error: %v", err)
	}
	if !reflect.DeepEqual(decoded, original) {
		t.Fatalf("Project round trip = %#v, want %#v", decoded, original)
	}
}

func TestProjectJSONLoadsDraftAndPersistedShapes(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		want        Project
		createdZero bool
		updatedZero bool
	}{
		{
			name:        "draft",
			payload:     `{"id":0,"connectionId":77,"name":"Draft Project","description":""}`,
			want:        Project{ID: 0, ConnectionID: 77, Name: "Draft Project", Description: ""},
			createdZero: true,
			updatedZero: true,
		},
		{
			name:    "persisted",
			payload: `{"id":88,"connectionId":77,"name":"Persisted Project","description":"Loaded from storage","createdAt":"2026-06-10T09:30:00Z","updatedAt":"2026-06-10T10:45:00Z"}`,
			want: Project{
				ID:           88,
				ConnectionID: 77,
				Name:         "Persisted Project",
				Description:  "Loaded from storage",
				CreatedAt:    time.Date(2026, 6, 10, 9, 30, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2026, 6, 10, 10, 45, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var decoded Project
			if err := json.Unmarshal([]byte(tt.payload), &decoded); err != nil {
				t.Fatalf("Unmarshal(Project) returned error: %v", err)
			}
			if !reflect.DeepEqual(decoded, tt.want) {
				t.Fatalf("decoded Project = %#v, want %#v", decoded, tt.want)
			}
			if decoded.CreatedAt.IsZero() != tt.createdZero {
				t.Fatalf("CreatedAt zero = %v, want %v", decoded.CreatedAt.IsZero(), tt.createdZero)
			}
			if decoded.UpdatedAt.IsZero() != tt.updatedZero {
				t.Fatalf("UpdatedAt zero = %v, want %v", decoded.UpdatedAt.IsZero(), tt.updatedZero)
			}
		})
	}
}

func TestProjectTableExposesStableContractFields(t *testing.T) {
	projectTableType := reflect.TypeOf(ProjectTable{})

	assertProjectJSONTags(t, projectTableType, map[string]string{
		"ID":             "id",
		"ProjectID":      "projectId",
		"TableID":        "tableId",
		"RowCount":       "rowCount",
		"TruncateBefore": "truncateBefore",
		"ExecutionOrder": "executionOrder",
		"CreatedAt":      "createdAt",
		"UpdatedAt":      "updatedAt",
	})
	assertProjectStructJSONFieldSet(t, projectTableType, []string{"id", "projectId", "tableId", "rowCount", "truncateBefore", "executionOrder", "createdAt", "updatedAt"})

	assertProjectFieldType(t, projectTableType, "ID", reflect.TypeOf(int64(0)))
	assertProjectFieldType(t, projectTableType, "ProjectID", reflect.TypeOf(int64(0)))
	assertProjectFieldType(t, projectTableType, "TableID", reflect.TypeOf(int64(0)))
	assertProjectFieldType(t, projectTableType, "RowCount", reflect.TypeOf((*int)(nil)))
	assertProjectFieldType(t, projectTableType, "TruncateBefore", reflect.TypeOf(false))
	assertProjectFieldType(t, projectTableType, "ExecutionOrder", reflect.TypeOf(0))
	assertProjectFieldType(t, projectTableType, "CreatedAt", reflect.TypeOf(time.Time{}))
	assertProjectFieldType(t, projectTableType, "UpdatedAt", reflect.TypeOf(time.Time{}))
}

func TestProjectTableJSONRoundTripPreservesNullableRowCountStates(t *testing.T) {
	createdAt := time.Date(2026, 6, 10, 11, 15, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 6, 10, 11, 45, 0, 0, time.UTC)
	zeroRows := 0
	positiveRows := 25

	tests := []struct {
		name            string
		rowCount        *int
		wantRawRowCount string
	}{
		{name: "nil means dynamically derived", rowCount: nil, wantRawRowCount: "null"},
		{name: "zero means explicitly generate no rows", rowCount: &zeroRows, wantRawRowCount: "0"},
		{name: "positive means explicit row target", rowCount: &positiveRows, wantRawRowCount: "25"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := ProjectTable{
				ID:             301,
				ProjectID:      101,
				TableID:        202,
				RowCount:       tt.rowCount,
				TruncateBefore: false,
				ExecutionOrder: 3,
				CreatedAt:      createdAt,
				UpdatedAt:      updatedAt,
			}

			encoded, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("Marshal(ProjectTable) returned error: %v", err)
			}

			var fields map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &fields); err != nil {
				t.Fatalf("Unmarshal encoded ProjectTable into field map returned error: %v", err)
			}
			assertProjectJSONFieldsPresent(t, fields, "id", "projectId", "tableId", "rowCount", "truncateBefore", "executionOrder", "createdAt", "updatedAt")
			assertProjectJSONFieldsAbsent(t, fields, "project_id", "table_id", "row_count", "truncate_before", "execution_order", "fieldRules", "generatorConfig", "executionStatus", "runtimeState", "relations")
			if got := string(fields["rowCount"]); got != tt.wantRawRowCount {
				t.Fatalf("encoded rowCount = %s, want %s", got, tt.wantRawRowCount)
			}

			var decoded ProjectTable
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("Unmarshal(ProjectTable) returned error: %v", err)
			}
			if !reflect.DeepEqual(decoded, original) {
				t.Fatalf("ProjectTable round trip = %#v, want %#v", decoded, original)
			}
		})
	}
}

func TestProjectTableLoadsPersistedProjectReferenceAndFalseTruncate(t *testing.T) {
	payload := `{"id":301,"projectId":101,"tableId":202,"rowCount":null,"truncateBefore":false,"executionOrder":3,"createdAt":"2026-06-10T11:15:00Z","updatedAt":"2026-06-10T11:45:00Z"}`

	var decoded ProjectTable
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		t.Fatalf("Unmarshal(ProjectTable) returned error: %v", err)
	}

	if decoded.ProjectID <= 0 {
		t.Fatalf("ProjectTable.ProjectID = %d, want positive persisted Project reference", decoded.ProjectID)
	}
	if decoded.RowCount != nil {
		t.Fatalf("ProjectTable.RowCount = %#v, want nil dynamic row count", decoded.RowCount)
	}
	if decoded.TruncateBefore {
		t.Fatalf("ProjectTable.TruncateBefore = true, want false preserved from JSON")
	}
	if decoded.ExecutionOrder != 3 {
		t.Fatalf("ProjectTable.ExecutionOrder = %d, want persisted snapshot 3", decoded.ExecutionOrder)
	}
}

func TestProjectTableExcludesFieldRulesRelationsAndRuntimeState(t *testing.T) {
	projectTableType := reflect.TypeOf(ProjectTable{})
	for _, forbidden := range []string{
		"TableRelationID",
		"ParentProjectTableID",
		"ChildProjectTableID",
		"MultiplierMin",
		"MultiplierMax",
		"RelValueSource",
		"RelSourceSQL",
		"GeneratorConfig",
		"FieldRules",
		"GenerationRules",
		"ExecutionStatus",
		"RuntimeState",
		"Status",
		"Relations",
		"RoleMatrix",
	} {
		if _, ok := projectTableType.FieldByName(forbidden); ok {
			t.Fatalf("ProjectTable exposes later-task or out-of-scope field %s", forbidden)
		}
	}
}

func TestProjectAggregateRootExcludesLaterTaskContracts(t *testing.T) {
	projectType := reflect.TypeOf(Project{})
	for _, forbidden := range []string{
		"ProjectID",
		"TableID",
		"RowCount",
		"TruncateBefore",
		"ExecutionOrder",
		"TableRelationID",
		"ParentProjectTableID",
		"ChildProjectTableID",
		"MultiplierMin",
		"MultiplierMax",
		"RelValueSource",
		"RelSourceSQL",
		"GeneratorConfig",
		"ExecutionStatus",
		"Status",
		"Tables",
		"Relations",
	} {
		if _, ok := projectType.FieldByName(forbidden); ok {
			t.Fatalf("Project exposes later-task or out-of-scope field %s", forbidden)
		}
	}
}

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

func TestProjectDomainExportsOnlyCurrentTaskContract(t *testing.T) {
	allowedExportedTypes := map[string]bool{
		"Project":      true,
		"ProjectTable": true,
	}

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
						if typedSpec.Name.IsExported() && !allowedExportedTypes[typedSpec.Name.Name] {
							t.Fatalf("%s exports %s outside ProjectModel task boundary", file, typedSpec.Name.Name)
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

func assertProjectJSONTags(t *testing.T, typ reflect.Type, expected map[string]string) {
	t.Helper()

	for fieldName, jsonName := range expected {
		field, ok := typ.FieldByName(fieldName)
		if !ok {
			t.Fatalf("%s missing field %s", typ.Name(), fieldName)
		}

		tag := field.Tag.Get("json")
		if tag == "" {
			t.Fatalf("%s.%s missing json tag", typ.Name(), fieldName)
		}
		actualName := strings.Split(tag, ",")[0]
		if actualName != jsonName {
			t.Fatalf("%s.%s json tag = %q, want %q", typ.Name(), fieldName, actualName, jsonName)
		}
		if strings.Contains(tag, "omitempty") {
			t.Fatalf("%s.%s must not use omitempty because Project fields are part of the stable contract", typ.Name(), fieldName)
		}
	}
}

func assertProjectStructJSONFieldSet(t *testing.T, typ reflect.Type, expected []string) {
	t.Helper()

	if typ.NumField() != len(expected) {
		t.Fatalf("%s field count = %d, want %d", typ.Name(), typ.NumField(), len(expected))
	}

	actual := make([]string, 0, typ.NumField())
	for i := range typ.NumField() {
		field := typ.Field(i)
		actual = append(actual, strings.Split(field.Tag.Get("json"), ",")[0])
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%s json fields = %#v, want %#v", typ.Name(), actual, expected)
	}
}

func assertProjectFieldType(t *testing.T, typ reflect.Type, fieldName string, expected reflect.Type) {
	t.Helper()

	field, ok := typ.FieldByName(fieldName)
	if !ok {
		t.Fatalf("%s missing field %s", typ.Name(), fieldName)
	}
	if field.Type != expected {
		t.Fatalf("%s.%s type = %v, want %v", typ.Name(), fieldName, field.Type, expected)
	}
}

func assertProjectJSONFieldsPresent(t *testing.T, fields map[string]json.RawMessage, expected ...string) {
	t.Helper()

	for _, field := range expected {
		if _, ok := fields[field]; !ok {
			t.Fatalf("encoded JSON missing stable field %q in %#v", field, fields)
		}
	}
}

func assertProjectJSONFieldsAbsent(t *testing.T, fields map[string]json.RawMessage, absent ...string) {
	t.Helper()

	for _, field := range absent {
		if _, ok := fields[field]; ok {
			t.Fatalf("encoded JSON contains out-of-contract field %q in %#v", field, fields)
		}
	}
}
