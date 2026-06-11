package execution

import (
	"encoding/json"
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

func TestExecutionStatusEnumsExposeStableStringsAndUnknownRecognition(t *testing.T) {
	taskStatuses := map[ExecutionTaskStatus]string{
		ExecutionTaskStatusRunning:       "RUNNING",
		ExecutionTaskStatusSuccess:       "SUCCESS",
		ExecutionTaskStatusPartialFailed: "PARTIAL_FAILED",
		ExecutionTaskStatusFailed:        "FAILED",
	}
	for status, want := range taskStatuses {
		if !status.IsKnown() {
			t.Fatalf("%s should be recognized as a known task status", status)
		}
		if status.IsUnknown() {
			t.Fatalf("%s should not be recognized as an unknown task status", status)
		}
		if got := status.String(); got != want {
			t.Fatalf("ExecutionTaskStatus.String() = %q, want %q", got, want)
		}
	}

	unknownTaskStatus := ExecutionTaskStatus("QUEUED")
	if unknownTaskStatus.IsKnown() {
		t.Fatalf("unknown task status %q should not be known", unknownTaskStatus)
	}
	if !unknownTaskStatus.IsUnknown() {
		t.Fatalf("unknown task status %q should be explicitly unknown", unknownTaskStatus)
	}
	if got := unknownTaskStatus.String(); got != "QUEUED" {
		t.Fatalf("unknown task status String() = %q, want preserved value", got)
	}

	tableStatuses := map[ExecutionTableStatus]string{
		ExecutionTableStatusPending: "PENDING",
		ExecutionTableStatusRunning: "RUNNING",
		ExecutionTableStatusSuccess: "SUCCESS",
		ExecutionTableStatusFailed:  "FAILED",
		ExecutionTableStatusSkipped: "SKIPPED",
	}
	for status, want := range tableStatuses {
		if !status.IsKnown() {
			t.Fatalf("%s should be recognized as a known table status", status)
		}
		if status.IsUnknown() {
			t.Fatalf("%s should not be recognized as an unknown table status", status)
		}
		if got := status.String(); got != want {
			t.Fatalf("ExecutionTableStatus.String() = %q, want %q", got, want)
		}
	}

	unknownTableStatus := ExecutionTableStatus("RETRYING")
	if unknownTableStatus.IsKnown() {
		t.Fatalf("unknown table status %q should not be known", unknownTableStatus)
	}
	if !unknownTableStatus.IsUnknown() {
		t.Fatalf("unknown table status %q should be explicitly unknown", unknownTableStatus)
	}
	if got := unknownTableStatus.String(); got != "RETRYING" {
		t.Fatalf("unknown table status String() = %q, want preserved value", got)
	}
}

func TestExecutionStatusJSONRoundTripPreservesKnownAndUnknownValues(t *testing.T) {
	taskStatusCases := []struct {
		name      string
		status    ExecutionTaskStatus
		jsonValue string
		known     bool
	}{
		{name: "running", status: ExecutionTaskStatusRunning, jsonValue: `"RUNNING"`, known: true},
		{name: "partial failed", status: ExecutionTaskStatusPartialFailed, jsonValue: `"PARTIAL_FAILED"`, known: true},
		{name: "unknown", status: ExecutionTaskStatus("QUEUED"), jsonValue: `"QUEUED"`, known: false},
	}
	for _, tt := range taskStatusCases {
		t.Run("task "+tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.status)
			if err != nil {
				t.Fatalf("Marshal(ExecutionTaskStatus) returned error: %v", err)
			}
			if string(encoded) != tt.jsonValue {
				t.Fatalf("Marshal(ExecutionTaskStatus) = %s, want %s", encoded, tt.jsonValue)
			}

			var decoded ExecutionTaskStatus
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("Unmarshal(ExecutionTaskStatus) returned error: %v", err)
			}
			if decoded != tt.status {
				t.Fatalf("decoded ExecutionTaskStatus = %q, want %q", decoded, tt.status)
			}
			if decoded.IsKnown() != tt.known {
				t.Fatalf("decoded IsKnown() = %v, want %v", decoded.IsKnown(), tt.known)
			}
		})
	}

	tableStatusCases := []struct {
		name      string
		status    ExecutionTableStatus
		jsonValue string
		known     bool
	}{
		{name: "pending", status: ExecutionTableStatusPending, jsonValue: `"PENDING"`, known: true},
		{name: "skipped", status: ExecutionTableStatusSkipped, jsonValue: `"SKIPPED"`, known: true},
		{name: "unknown", status: ExecutionTableStatus("RETRYING"), jsonValue: `"RETRYING"`, known: false},
	}
	for _, tt := range tableStatusCases {
		t.Run("table "+tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.status)
			if err != nil {
				t.Fatalf("Marshal(ExecutionTableStatus) returned error: %v", err)
			}
			if string(encoded) != tt.jsonValue {
				t.Fatalf("Marshal(ExecutionTableStatus) = %s, want %s", encoded, tt.jsonValue)
			}

			var decoded ExecutionTableStatus
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("Unmarshal(ExecutionTableStatus) returned error: %v", err)
			}
			if decoded != tt.status {
				t.Fatalf("decoded ExecutionTableStatus = %q, want %q", decoded, tt.status)
			}
			if decoded.IsKnown() != tt.known {
				t.Fatalf("decoded IsKnown() = %v, want %v", decoded.IsKnown(), tt.known)
			}
		})
	}
}

func TestExecutionStatusJSONRejectsNonStringValues(t *testing.T) {
	for _, raw := range []string{`1`, `true`, `{"status":"RUNNING"}`} {
		t.Run("task "+raw, func(t *testing.T) {
			var decoded ExecutionTaskStatus
			if err := json.Unmarshal([]byte(raw), &decoded); err == nil {
				t.Fatalf("Unmarshal ExecutionTaskStatus should reject non-string JSON %s", raw)
			}
		})

		t.Run("table "+raw, func(t *testing.T) {
			var decoded ExecutionTableStatus
			if err := json.Unmarshal([]byte(raw), &decoded); err == nil {
				t.Fatalf("Unmarshal ExecutionTableStatus should reject non-string JSON %s", raw)
			}
		})
	}
}

func TestExecutionTableResultAndErrorSnapshotJSONContract(t *testing.T) {
	tableID := int64(300)
	occurredAt := time.Date(2026, 6, 11, 9, 30, 0, 0, time.UTC)
	createdAt := time.Date(2026, 6, 11, 9, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 6, 11, 9, 45, 0, 0, time.UTC)
	result := ExecutionTableResult{
		ID:                 20,
		ExecutionTaskID:    10,
		TableID:            &tableID,
		TableNameSnapshot:  "orders",
		SchemaNameSnapshot: "public",
		RowsWritten:        42,
		Status:             ExecutionTableStatusFailed,
		ErrorSnapshot: &ExecutionErrorSnapshot{
			Code:       "WRITE_FAILED",
			Message:    "failed to write generated rows",
			FieldPath:  "tableResults[0]",
			OccurredAt: occurredAt,
		},
		ExecutionOrder: 3,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}

	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal(ExecutionTableResult) returned error: %v", err)
	}
	want := `{"id":20,"executionTaskId":10,"tableId":300,"tableNameSnapshot":"orders","schemaNameSnapshot":"public","rowsWritten":42,"status":"FAILED","errorSnapshot":{"code":"WRITE_FAILED","message":"failed to write generated rows","fieldPath":"tableResults[0]","occurredAt":"2026-06-11T09:30:00Z"},"executionOrder":3,"createdAt":"2026-06-11T09:00:00Z","updatedAt":"2026-06-11T09:45:00Z"}`
	if string(encoded) != want {
		t.Fatalf("ExecutionTableResult JSON = %s, want %s", encoded, want)
	}

	var decoded ExecutionTableResult
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(ExecutionTableResult) returned error: %v", err)
	}
	if decoded.TableID == nil || *decoded.TableID != tableID {
		t.Fatalf("decoded TableID = %#v, want %d", decoded.TableID, tableID)
	}
	if decoded.Status != ExecutionTableStatusFailed || !decoded.Status.IsKnown() {
		t.Fatalf("decoded Status = %q, want known FAILED", decoded.Status)
	}
	if decoded.ErrorSnapshot == nil || decoded.ErrorSnapshot.Message != "failed to write generated rows" {
		t.Fatalf("decoded ErrorSnapshot = %#v, want safe error snapshot", decoded.ErrorSnapshot)
	}
}

func TestExecutionTableResultPreservesUnknownStatusAndOmitEmptyOptionalFields(t *testing.T) {
	raw := `{"id":21,"executionTaskId":10,"tableNameSnapshot":"deleted_table","schemaNameSnapshot":"archive","rowsWritten":0,"status":"RETRYING","executionOrder":4,"createdAt":"2026-06-11T09:00:00Z","updatedAt":"2026-06-11T09:00:00Z"}`

	var decoded ExecutionTableResult
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		t.Fatalf("Unmarshal(ExecutionTableResult with unknown status) returned error: %v", err)
	}
	if decoded.TableID != nil {
		t.Fatalf("decoded TableID = %#v, want nil for deleted table history", decoded.TableID)
	}
	if decoded.ErrorSnapshot != nil {
		t.Fatalf("decoded ErrorSnapshot = %#v, want nil when omitted", decoded.ErrorSnapshot)
	}
	if !decoded.Status.IsUnknown() || decoded.Status.String() != "RETRYING" {
		t.Fatalf("decoded Status = %q, want preserved unknown RETRYING", decoded.Status)
	}

	encoded, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("Marshal(ExecutionTableResult with unknown status) returned error: %v", err)
	}
	if strings.Contains(string(encoded), "tableId") {
		t.Fatalf("ExecutionTableResult JSON = %s, want omitted tableId", encoded)
	}
	if strings.Contains(string(encoded), "errorSnapshot") {
		t.Fatalf("ExecutionTableResult JSON = %s, want omitted errorSnapshot", encoded)
	}
	if !strings.Contains(string(encoded), `"status":"RETRYING"`) {
		t.Fatalf("ExecutionTableResult JSON = %s, want preserved unknown status", encoded)
	}
}

func TestExecutionValidationIssueContract(t *testing.T) {
	assertExecutionJSONTags(t, reflect.TypeOf(ValidationIssue{}), map[string]string{
		"Path":    "path",
		"Code":    "code",
		"Message": "message",
	})
	assertExecutionStructJSONFieldSet(t, reflect.TypeOf(ValidationIssue{}), []string{"path", "code", "message"})
	assertExecutionFieldType(t, reflect.TypeOf(ValidationIssue{}), "Path", reflect.TypeOf(""))
	assertExecutionFieldType(t, reflect.TypeOf(ValidationIssue{}), "Code", reflect.TypeOf(ValidationIssueCode("")))
	assertExecutionFieldType(t, reflect.TypeOf(ValidationIssue{}), "Message", reflect.TypeOf(""))

	issue := ValidationIssue{
		Path:    "tableResults[0].rowsWritten",
		Code:    ValidationIssueCodeInvalidRange,
		Message: "rowsWritten must be greater than or equal to zero",
	}
	encoded, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("Marshal(ValidationIssue) returned error: %v", err)
	}
	want := `{"path":"tableResults[0].rowsWritten","code":"INVALID_RANGE","message":"rowsWritten must be greater than or equal to zero"}`
	if string(encoded) != want {
		t.Fatalf("ValidationIssue JSON = %s, want %s", encoded, want)
	}

	var decoded ValidationIssue
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(ValidationIssue) returned error: %v", err)
	}
	if decoded != issue {
		t.Fatalf("decoded ValidationIssue = %#v, want %#v", decoded, issue)
	}
}

func TestExecutionValidationIssueCodesAreStableAndClassified(t *testing.T) {
	codes := map[ValidationIssueCode]string{
		ValidationIssueCodeRequired:                 "REQUIRED",
		ValidationIssueCodeTooLong:                  "TOO_LONG",
		ValidationIssueCodeInvalidEnum:              "INVALID_ENUM",
		ValidationIssueCodeInvalidRange:             "INVALID_RANGE",
		ValidationIssueCodeInvalidReference:         "INVALID_REFERENCE",
		ValidationIssueCodeInvalidTimeRange:         "INVALID_TIME_RANGE",
		ValidationIssueCodeInvalidNestedModel:       "INVALID_NESTED_MODEL",
		ValidationIssueCodeSensitiveValueNotAllowed: "SENSITIVE_VALUE_NOT_ALLOWED",
	}
	for code, want := range codes {
		if got := string(code); got != want {
			t.Fatalf("ValidationIssueCode value = %q, want %q", got, want)
		}
		if !code.IsKnown() {
			t.Fatalf("%s should be recognized as a known validation issue code", code)
		}
		if code.IsUnknown() {
			t.Fatalf("%s should not be recognized as an unknown validation issue code", code)
		}
		if got := code.String(); got != want {
			t.Fatalf("ValidationIssueCode.String() = %q, want %q", got, want)
		}
	}

	unknown := ValidationIssueCode("FUTURE_CODE")
	if unknown.IsKnown() {
		t.Fatalf("unknown validation issue code %q should not be known", unknown)
	}
	if !unknown.IsUnknown() {
		t.Fatalf("unknown validation issue code %q should be explicitly unknown", unknown)
	}
	if got := unknown.String(); got != "FUTURE_CODE" {
		t.Fatalf("unknown validation issue code String() = %q, want preserved value", got)
	}
}

func TestExecutionValidationResultSupportsMultipleIssues(t *testing.T) {
	result := ValidationResult{}
	if result.HasIssues() {
		t.Fatalf("empty ValidationResult should not have issues")
	}

	result.AddIssue("projectId", ValidationIssueCodeRequired, "projectId is required")
	result.AddIssue("tableResults[0].executionTaskId", ValidationIssueCodeInvalidReference, "executionTaskId must reference the task")
	if !result.HasIssues() {
		t.Fatalf("ValidationResult with issues should report issues")
	}
	if len(result.Issues) != 2 {
		t.Fatalf("ValidationResult issues length = %d, want 2", len(result.Issues))
	}
	if result.Issues[0].Path != "projectId" || result.Issues[1].Path != "tableResults[0].executionTaskId" {
		t.Fatalf("ValidationResult issues = %#v, want field paths preserved in order", result.Issues)
	}

	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal(ValidationResult) returned error: %v", err)
	}
	want := `{"issues":[{"path":"projectId","code":"REQUIRED","message":"projectId is required"},{"path":"tableResults[0].executionTaskId","code":"INVALID_REFERENCE","message":"executionTaskId must reference the task"}]}`
	if string(encoded) != want {
		t.Fatalf("ValidationResult JSON = %s, want %s", encoded, want)
	}
}

func TestExecutionValidationIssueExposesOnlySafeMessageContract(t *testing.T) {
	typ := reflect.TypeOf(ValidationIssue{})
	if typ.NumField() != 3 {
		t.Fatalf("ValidationIssue field count = %d, want only path, code, and safe message", typ.NumField())
	}
	for _, disallowed := range []string{"RawValue", "Value", "Details", "Cause", "Credential", "SQL", "GeneratedData"} {
		if _, ok := typ.FieldByName(disallowed); ok {
			t.Fatalf("ValidationIssue must not expose unsafe diagnostic field %s", disallowed)
		}
	}

	issue := ValidationIssue{Path: "errorSnapshot.message", Code: ValidationIssueCodeSensitiveValueNotAllowed, Message: "message must not contain sensitive values"}
	encoded, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("Marshal(ValidationIssue) returned error: %v", err)
	}
	jsonText := string(encoded)
	for _, forbidden := range []string{"password", "credential", "select *", "generatedData"} {
		if strings.Contains(strings.ToLower(jsonText), strings.ToLower(forbidden)) {
			t.Fatalf("ValidationIssue JSON %s should not contain unsafe diagnostic content %q", jsonText, forbidden)
		}
	}
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
