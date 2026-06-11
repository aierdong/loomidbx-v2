package execution

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestExecutionDomainScaffoldExportsStableShapes(t *testing.T) {
	assertExecutionJSONTags(t, reflect.TypeOf(GenerationJob{}), map[string]string{
		"Task":         "task",
		"TableResults": "tableResults",
	})
	assertExecutionStructJSONFieldSet(t, reflect.TypeOf(GenerationJob{}), []string{"task", "tableResults"})

	assertExecutionJSONTags(t, reflect.TypeOf(ExecutionTask{}), map[string]string{
		"ID":        "id",
		"ProjectID": "projectId",
		"TaskName":  "taskName",
		"Status":    "status",
		"StartedAt": "startedAt",
		"EndedAt":   "endedAt,omitempty",
		"CreatedAt": "createdAt",
	})
	assertExecutionStructJSONFieldSet(t, reflect.TypeOf(ExecutionTask{}), []string{"id", "projectId", "taskName", "status", "startedAt", "endedAt", "createdAt"})

	assertExecutionJSONTags(t, reflect.TypeOf(ExecutionTableResult{}), map[string]string{
		"ID":                 "id",
		"ExecutionTaskID":    "executionTaskId",
		"TableID":            "tableId,omitempty",
		"TableNameSnapshot":  "tableNameSnapshot",
		"SchemaNameSnapshot": "schemaNameSnapshot",
		"RowsWritten":        "rowsWritten",
		"Status":             "status",
		"ErrorSnapshot":      "errorSnapshot,omitempty",
		"ExecutionOrder":     "executionOrder",
		"CreatedAt":          "createdAt",
		"UpdatedAt":          "updatedAt",
	})
	assertExecutionStructJSONFieldSet(t, reflect.TypeOf(ExecutionTableResult{}), []string{"id", "executionTaskId", "tableId", "tableNameSnapshot", "schemaNameSnapshot", "rowsWritten", "status", "errorSnapshot", "executionOrder", "createdAt", "updatedAt"})

	assertExecutionJSONTags(t, reflect.TypeOf(ExecutionErrorSnapshot{}), map[string]string{
		"Code":       "code",
		"Message":    "message",
		"FieldPath":  "fieldPath,omitempty",
		"OccurredAt": "occurredAt",
	})
	assertExecutionStructJSONFieldSet(t, reflect.TypeOf(ExecutionErrorSnapshot{}), []string{"code", "message", "fieldPath", "occurredAt"})

	assertExecutionFieldType(t, reflect.TypeOf(GenerationJob{}), "Task", reflect.TypeOf(ExecutionTask{}))
	assertExecutionFieldType(t, reflect.TypeOf(GenerationJob{}), "TableResults", reflect.TypeOf([]ExecutionTableResult{}))
	assertExecutionFieldType(t, reflect.TypeOf(ExecutionTask{}), "ID", reflect.TypeOf(int64(0)))
	assertExecutionFieldType(t, reflect.TypeOf(ExecutionTask{}), "ProjectID", reflect.TypeOf(int64(0)))
	assertExecutionFieldType(t, reflect.TypeOf(ExecutionTask{}), "TaskName", reflect.TypeOf(""))
	assertExecutionFieldType(t, reflect.TypeOf(ExecutionTask{}), "Status", reflect.TypeOf(ExecutionTaskStatus("")))
	assertExecutionFieldType(t, reflect.TypeOf(ExecutionTask{}), "StartedAt", reflect.TypeOf(time.Time{}))
	assertExecutionFieldType(t, reflect.TypeOf(ExecutionTask{}), "EndedAt", reflect.TypeOf((*time.Time)(nil)))
	assertExecutionFieldType(t, reflect.TypeOf(ExecutionTask{}), "CreatedAt", reflect.TypeOf(time.Time{}))
	assertExecutionFieldType(t, reflect.TypeOf(ExecutionTableResult{}), "ErrorSnapshot", reflect.TypeOf((*ExecutionErrorSnapshot)(nil)))
}

func TestExecutionDomainScaffoldDeclaresStableEnums(t *testing.T) {
	taskStatuses := map[ExecutionTaskStatus]string{
		ExecutionTaskStatusRunning:       "RUNNING",
		ExecutionTaskStatusSuccess:       "SUCCESS",
		ExecutionTaskStatusPartialFailed: "PARTIAL_FAILED",
		ExecutionTaskStatusFailed:        "FAILED",
	}
	for status, want := range taskStatuses {
		if got := string(status); got != want {
			t.Fatalf("ExecutionTaskStatus value = %q, want %q", got, want)
		}
	}

	tableStatuses := map[ExecutionTableStatus]string{
		ExecutionTableStatusPending: "PENDING",
		ExecutionTableStatusRunning: "RUNNING",
		ExecutionTableStatusSuccess: "SUCCESS",
		ExecutionTableStatusFailed:  "FAILED",
		ExecutionTableStatusSkipped: "SKIPPED",
	}
	for status, want := range tableStatuses {
		if got := string(status); got != want {
			t.Fatalf("ExecutionTableStatus value = %q, want %q", got, want)
		}
	}

	var _ GenerationJob
	var _ ExecutionTask
	var _ ExecutionTableResult
	var _ ExecutionErrorSnapshot
	var _ ExecutionTaskStatus
	var _ ExecutionTableStatus
}

func assertExecutionJSONTags(t *testing.T, typ reflect.Type, expected map[string]string) {
	t.Helper()
	for fieldName, want := range expected {
		field, ok := typ.FieldByName(fieldName)
		if !ok {
			t.Fatalf("%s.%s field is missing", typ.Name(), fieldName)
		}
		if got := field.Tag.Get("json"); got != want {
			t.Fatalf("%s.%s json tag = %q, want %q", typ.Name(), fieldName, got, want)
		}
	}
}

func assertExecutionStructJSONFieldSet(t *testing.T, typ reflect.Type, want []string) {
	t.Helper()
	if typ.NumField() != len(want) {
		t.Fatalf("%s field count = %d, want %d", typ.Name(), typ.NumField(), len(want))
	}
	for index, wantJSON := range want {
		field := typ.Field(index)
		got := field.Tag.Get("json")
		if got == wantJSON {
			continue
		}
		if commaIndex := strings.IndexByte(got, ','); commaIndex >= 0 && got[:commaIndex] == wantJSON {
			continue
		}
		t.Fatalf("%s field %d json tag = %q, want field name %q", typ.Name(), index, got, wantJSON)
	}
}

func assertExecutionFieldType(t *testing.T, typ reflect.Type, fieldName string, want reflect.Type) {
	t.Helper()
	field, ok := typ.FieldByName(fieldName)
	if !ok {
		t.Fatalf("%s.%s field is missing", typ.Name(), fieldName)
	}
	if field.Type != want {
		t.Fatalf("%s.%s type = %s, want %s", typ.Name(), fieldName, field.Type, want)
	}
}
