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

func TestRelationValueSourceExposesStableEnumStringsAndUnknownRecognition(t *testing.T) {
	tests := []struct {
		name   string
		source RelationValueSource
		want   string
		known  bool
	}{
		{name: "from execution", source: RelationValueSourceFromExecution, want: "FROM_EXECUTION", known: true},
		{name: "from db query", source: RelationValueSourceFromDBQuery, want: "FROM_DB_QUERY", known: true},
		{name: "merged", source: RelationValueSourceMerged, want: "MERGED", known: true},
		{name: "unknown", source: RelationValueSource("ARCHIVED"), want: "ARCHIVED", known: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.source); got != tt.want {
				t.Fatalf("RelationValueSource string = %q, want %q", got, tt.want)
			}
			if got := tt.source.IsKnown(); got != tt.known {
				t.Fatalf("RelationValueSource(%q).IsKnown() = %v, want %v", tt.source, got, tt.known)
			}

			encoded, err := json.Marshal(tt.source)
			if err != nil {
				t.Fatalf("Marshal(RelationValueSource) returned error: %v", err)
			}
			if got := string(encoded); got != `"`+tt.want+`"` {
				t.Fatalf("encoded RelationValueSource = %s, want %q", got, tt.want)
			}

			var decoded RelationValueSource
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("Unmarshal(RelationValueSource) returned error: %v", err)
			}
			if decoded != tt.source {
				t.Fatalf("decoded RelationValueSource = %q, want %q", decoded, tt.source)
			}
		})
	}
}

func TestProjectTableRelationExposesStableContractFields(t *testing.T) {
	relationType := reflect.TypeOf(ProjectTableRelation{})

	assertProjectJSONTags(t, relationType, map[string]string{
		"ID":                   "id",
		"ProjectID":            "projectId",
		"TableRelationID":      "tableRelationId",
		"ParentProjectTableID": "parentProjectTableId",
		"ChildProjectTableID":  "childProjectTableId",
		"MultiplierMin":        "multiplierMin",
		"MultiplierMax":        "multiplierMax",
		"RelValueSource":       "relValueSource",
		"RelSourceSQL":         "relSourceSql",
		"CreatedAt":            "createdAt",
		"UpdatedAt":            "updatedAt",
	})
	assertProjectStructJSONFieldSet(t, relationType, []string{"id", "projectId", "tableRelationId", "parentProjectTableId", "childProjectTableId", "multiplierMin", "multiplierMax", "relValueSource", "relSourceSql", "createdAt", "updatedAt"})

	assertProjectFieldType(t, relationType, "ID", reflect.TypeOf(int64(0)))
	assertProjectFieldType(t, relationType, "ProjectID", reflect.TypeOf(int64(0)))
	assertProjectFieldType(t, relationType, "TableRelationID", reflect.TypeOf(int64(0)))
	assertProjectFieldType(t, relationType, "ParentProjectTableID", reflect.TypeOf((*int64)(nil)))
	assertProjectFieldType(t, relationType, "ChildProjectTableID", reflect.TypeOf(int64(0)))
	assertProjectFieldType(t, relationType, "MultiplierMin", reflect.TypeOf(0))
	assertProjectFieldType(t, relationType, "MultiplierMax", reflect.TypeOf(0))
	assertProjectFieldType(t, relationType, "RelValueSource", reflect.TypeOf(RelationValueSource("")))
	assertProjectFieldType(t, relationType, "RelSourceSQL", reflect.TypeOf(""))
	assertProjectFieldType(t, relationType, "CreatedAt", reflect.TypeOf(time.Time{}))
	assertProjectFieldType(t, relationType, "UpdatedAt", reflect.TypeOf(time.Time{}))
}

func TestProjectTableRelationJSONRoundTripPreservesRelationSnapshot(t *testing.T) {
	createdAt := time.Date(2026, 6, 10, 12, 15, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 6, 10, 12, 45, 0, 0, time.UTC)
	parentID := int64(501)

	tests := []struct {
		name          string
		parentID      *int64
		source        RelationValueSource
		sourceSQL     string
		wantRawParent string
	}{
		{name: "parent omitted and values from db query", parentID: nil, source: RelationValueSourceFromDBQuery, sourceSQL: "select id from parent", wantRawParent: "null"},
		{name: "parent present and values from execution", parentID: &parentID, source: RelationValueSourceFromExecution, sourceSQL: "", wantRawParent: "501"},
		{name: "parent present and values merged", parentID: &parentID, source: RelationValueSourceMerged, sourceSQL: "select id from parent", wantRawParent: "501"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := ProjectTableRelation{
				ID:                   701,
				ProjectID:            101,
				TableRelationID:      401,
				ParentProjectTableID: tt.parentID,
				ChildProjectTableID:  502,
				MultiplierMin:        0,
				MultiplierMax:        3,
				RelValueSource:       tt.source,
				RelSourceSQL:         tt.sourceSQL,
				CreatedAt:            createdAt,
				UpdatedAt:            updatedAt,
			}

			encoded, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("Marshal(ProjectTableRelation) returned error: %v", err)
			}

			var fields map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &fields); err != nil {
				t.Fatalf("Unmarshal encoded ProjectTableRelation into field map returned error: %v", err)
			}
			assertProjectJSONFieldsPresent(t, fields, "id", "projectId", "tableRelationId", "parentProjectTableId", "childProjectTableId", "multiplierMin", "multiplierMax", "relValueSource", "relSourceSql", "createdAt", "updatedAt")
			assertProjectJSONFieldsAbsent(t, fields, "project_id", "table_relation_id", "parent_project_table_id", "child_project_table_id", "multiplier_min", "multiplier_max", "rel_value_source", "rel_source_sql", "relSourceSQL", "executionStatus", "runtimeState", "sqlResult", "generatedRows")
			if got := string(fields["parentProjectTableId"]); got != tt.wantRawParent {
				t.Fatalf("encoded parentProjectTableId = %s, want %s", got, tt.wantRawParent)
			}
			if got := string(fields["relValueSource"]); got != `"`+string(tt.source)+`"` {
				t.Fatalf("encoded relValueSource = %s, want %q", got, tt.source)
			}

			var decoded ProjectTableRelation
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("Unmarshal(ProjectTableRelation) returned error: %v", err)
			}
			if !reflect.DeepEqual(decoded, original) {
				t.Fatalf("ProjectTableRelation round trip = %#v, want %#v", decoded, original)
			}
		})
	}
}

func TestProjectTableRelationLoadsUnknownValueSourceWithoutExecutingSQL(t *testing.T) {
	payload := `{"id":701,"projectId":101,"tableRelationId":401,"parentProjectTableId":null,"childProjectTableId":502,"multiplierMin":1,"multiplierMax":2,"relValueSource":"ARCHIVED","relSourceSql":"select secret from parent","createdAt":"2026-06-10T12:15:00Z","updatedAt":"2026-06-10T12:45:00Z"}`

	var decoded ProjectTableRelation
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		t.Fatalf("Unmarshal(ProjectTableRelation) with unknown relValueSource returned error: %v", err)
	}

	if decoded.ProjectID <= 0 {
		t.Fatalf("ProjectTableRelation.ProjectID = %d, want positive persisted Project reference", decoded.ProjectID)
	}
	if decoded.ParentProjectTableID != nil {
		t.Fatalf("ParentProjectTableID = %#v, want nil for absent upstream ProjectTable", decoded.ParentProjectTableID)
	}
	if decoded.RelValueSource.IsKnown() {
		t.Fatalf("unknown relValueSource %q was reported as known", decoded.RelValueSource)
	}
	if decoded.RelSourceSQL != "select secret from parent" {
		t.Fatalf("RelSourceSQL = %q, want SQL text preserved as configuration snapshot", decoded.RelSourceSQL)
	}
}

func TestProjectValidationIssueCodeAndSeverityContractsAreStable(t *testing.T) {
	codeTests := []struct {
		name string
		code ProjectIssueCode
		want string
	}{
		{name: "validation failed", code: ProjectIssueCodeValidationFailed, want: "VALIDATION_FAILED"},
		{name: "required", code: ProjectIssueCodeRequired, want: "REQUIRED"},
		{name: "invalid id", code: ProjectIssueCodeInvalidID, want: "INVALID_ID"},
		{name: "invalid range", code: ProjectIssueCodeInvalidRange, want: "INVALID_RANGE"},
		{name: "invalid enum", code: ProjectIssueCodeInvalidEnum, want: "INVALID_ENUM"},
		{name: "invalid time", code: ProjectIssueCodeInvalidTime, want: "INVALID_TIME"},
		{name: "duplicate table", code: ProjectIssueCodeDuplicateTable, want: "DUPLICATE_TABLE"},
		{name: "sql required", code: ProjectIssueCodeSQLRequired, want: "SQL_REQUIRED"},
		{name: "parent required", code: ProjectIssueCodeParentRequired, want: "PARENT_REQUIRED"},
		{name: "out of scope", code: ProjectIssueCodeOutOfScope, want: "OUT_OF_SCOPE"},
	}

	for _, tt := range codeTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.code); got != tt.want {
				t.Fatalf("ProjectIssueCode string = %q, want %q", got, tt.want)
			}
			if !tt.code.IsKnown() {
				t.Fatalf("ProjectIssueCode(%q).IsKnown() = false, want true", tt.code)
			}
		})
	}
	if ProjectIssueCode("FUTURE_CODE").IsKnown() {
		t.Fatalf("unknown ProjectIssueCode should not be known")
	}

	severityTests := []struct {
		severity ProjectIssueSeverity
		want     string
	}{
		{severity: ProjectIssueSeverityInfo, want: "info"},
		{severity: ProjectIssueSeverityWarning, want: "warning"},
		{severity: ProjectIssueSeverityError, want: "error"},
	}
	for _, tt := range severityTests {
		if got := string(tt.severity); got != tt.want {
			t.Fatalf("ProjectIssueSeverity string = %q, want %q", got, tt.want)
		}
		if !tt.severity.IsKnown() {
			t.Fatalf("ProjectIssueSeverity(%q).IsKnown() = false, want true", tt.severity)
		}
	}
	if ProjectIssueSeverity("fatal").IsKnown() {
		t.Fatalf("unknown ProjectIssueSeverity should not be known")
	}
}

func TestProjectValidationIssueJSONShapeAndRoundTrip(t *testing.T) {
	issue := NewProjectValidationIssue("relations[0].relSourceSql", ProjectIssueCodeSQLRequired)

	encoded, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("Marshal(ProjectValidationIssue) returned error: %v", err)
	}

	const expected = `{"path":"relations[0].relSourceSql","code":"SQL_REQUIRED","severity":"error","message":"relSourceSql is required for the chosen relation value source"}`
	if string(encoded) != expected {
		t.Fatalf("ProjectValidationIssue JSON = %s, want stable lower camelCase shape %s", encoded, expected)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("Unmarshal encoded ProjectValidationIssue into map returned error: %v", err)
	}
	if got, want := len(fields), 4; got != want {
		t.Fatalf("encoded issue has %d fields, want exactly %d: %v", got, want, fields)
	}
	assertProjectJSONFieldsPresent(t, fields, "path", "code", "severity", "message")
	assertProjectJSONFieldsAbsent(t, fields, "field", "errorCode", "safeMessage", "sql", "relSourceSql")

	var decoded ProjectValidationIssue
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(ProjectValidationIssue) returned error: %v", err)
	}
	if !reflect.DeepEqual(decoded, issue) {
		t.Fatalf("decoded ProjectValidationIssue = %#v, want %#v", decoded, issue)
	}
}

func TestProjectValidationIssuesCanReturnMultipleProblemsAtOnce(t *testing.T) {
	issues := ProjectValidationIssues{
		NewProjectValidationIssue("project.name", ProjectIssueCodeRequired),
		NewProjectValidationIssue("tables[0].tableId", ProjectIssueCodeInvalidID),
		NewProjectValidationIssue("relations[0].relSourceSql", ProjectIssueCodeSQLRequired),
	}

	encoded, err := json.Marshal(issues)
	if err != nil {
		t.Fatalf("Marshal(ProjectValidationIssues) returned error: %v", err)
	}

	var decoded []ProjectValidationIssue
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(ProjectValidationIssues) returned error: %v", err)
	}
	if got, want := len(decoded), 3; got != want {
		t.Fatalf("decoded issue count = %d, want %d: %s", got, want, encoded)
	}
	assertProjectValidationIssuePaths(t, decoded, []string{"project.name", "tables[0].tableId", "relations[0].relSourceSql"})
	for _, issue := range decoded {
		if issue.Severity != ProjectIssueSeverityError {
			t.Fatalf("issue severity = %q, want error: %#v", issue.Severity, issue)
		}
		if strings.TrimSpace(issue.Message) == "" {
			t.Fatalf("issue message should be safe and non-empty: %#v", issue)
		}
	}
}

func TestProjectValidationIssueMessagesDoNotEchoSensitiveSQL(t *testing.T) {
	sensitiveSQL := "select password, token from customer_secret where api_key = 'secret'"
	issue := NewProjectValidationIssue("relations[0].relSourceSql", ProjectIssueCodeSQLRequired)

	encoded, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("Marshal(ProjectValidationIssue) returned error: %v", err)
	}
	lowerMessage := strings.ToLower(issue.Message)
	lowerEncoded := strings.ToLower(string(encoded))
	for _, leaked := range []string{"select", "password", "token", "customer_secret", "api_key", sensitiveSQL} {
		if strings.Contains(lowerMessage, strings.ToLower(leaked)) || strings.Contains(lowerEncoded, strings.ToLower(leaked)) {
			t.Fatalf("Project validation issue leaked SQL-sensitive content %q in %s", leaked, encoded)
		}
	}
}

func TestProjectTableRelationExcludesExecutionAlgorithmsAndRuntimeState(t *testing.T) {
	relationType := reflect.TypeOf(ProjectTableRelation{})
	for _, forbidden := range []string{
		"GeneratedRows",
		"WrittenRows",
		"ExecutionStatus",
		"RuntimeState",
		"Status",
		"SQLResult",
		"QueryResult",
		"DatabaseHandle",
		"Connection",
		"GeneratorConfig",
		"FieldRules",
		"GenerationRules",
		"SortOrder",
		"ResolvedParentRows",
	} {
		if _, ok := relationType.FieldByName(forbidden); ok {
			t.Fatalf("ProjectTableRelation exposes runtime, SQL execution, or out-of-scope field %s", forbidden)
		}
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
		"Project":                 true,
		"ProjectTable":            true,
		"ProjectTableRelation":    true,
		"RelationValueSource":     true,
		"ProjectIssueCode":        true,
		"ProjectIssueSeverity":    true,
		"ProjectValidationIssue":  true,
		"ProjectValidationIssues": true,
	}
	allowedExportedValues := map[string]bool{
		"RelationValueSourceFromExecution": true,
		"RelationValueSourceFromDBQuery":   true,
		"RelationValueSourceMerged":        true,
		"ProjectIssueCodeValidationFailed": true,
		"ProjectIssueCodeRequired":         true,
		"ProjectIssueCodeInvalidID":        true,
		"ProjectIssueCodeInvalidRange":     true,
		"ProjectIssueCodeInvalidEnum":      true,
		"ProjectIssueCodeInvalidTime":      true,
		"ProjectIssueCodeDuplicateTable":   true,
		"ProjectIssueCodeSQLRequired":      true,
		"ProjectIssueCodeParentRequired":   true,
		"ProjectIssueCodeOutOfScope":       true,
		"ProjectIssueSeverityInfo":         true,
		"ProjectIssueSeverityWarning":      true,
		"ProjectIssueSeverityError":        true,
	}
	allowedExportedFuncs := map[string]bool{
		"IsKnown":                   true,
		"NewProjectValidationIssue": true,
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
							if name.IsExported() && !allowedExportedValues[name.Name] {
								t.Fatalf("%s exports %s outside ProjectRelationModel or RelationValueSource task boundary", file, name.Name)
							}
						}
					}
				}
			case *ast.FuncDecl:
				if typed.Name.IsExported() && !allowedExportedFuncs[typed.Name.Name] {
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

func assertProjectValidationIssuePaths(t *testing.T, issues []ProjectValidationIssue, expected []string) {
	t.Helper()

	actual := make([]string, 0, len(issues))
	for _, issue := range issues {
		actual = append(actual, issue.Path)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("issue paths = %#v, want %#v in %#v", actual, expected, issues)
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
