package lifecycle

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestLifecycleErrorPublicShapeContainsOnlySafeSummaryFields(t *testing.T) {
	errorType := reflect.TypeOf(LifecycleError{})
	wantFields := []string{"Code", "Stage", "FieldPath", "SafeMessage"}
	if errorType.NumField() != len(wantFields) {
		t.Fatalf("LifecycleError has %d public fields, want only safe summary fields %v", errorType.NumField(), wantFields)
	}
	for index, want := range wantFields {
		field := errorType.Field(index)
		if field.Name != want {
			t.Fatalf("LifecycleError field[%d] = %s, want %s", index, field.Name, want)
		}
	}
}

func TestMapExecutionInputFromGenerationJobReturnsLifecycleSafeErrors(t *testing.T) {
	job := validGenerationJobSnapshot()
	job.Task.ID = 0
	job.TableResults[0].SchemaNameSnapshot = ""

	input, precheck := MapExecutionInputFromGenerationJob(job)

	if input != nil {
		t.Fatalf("expected no runnable input, got %#v", input)
	}
	if precheck.Passed {
		t.Fatal("expected invalid input boundary to fail precheck")
	}
	assertLifecycleError(t, precheck.BlockingErrors, LifecycleStageInputValidation, "task.id", LifecycleErrorCodeRequired)
	assertLifecycleError(t, precheck.BlockingErrors, LifecycleStageInputValidation, "tableResults[0].schemaNameSnapshot", LifecycleErrorCodeRequired)
}

func TestMapStateConflictErrorProducesSafeSummary(t *testing.T) {
	conflict := MapStateConflictError(LifecycleStageStateTransition, "state", "COMPLETED", "RUNNING")

	if conflict.Code != LifecycleErrorCodeStateConflict {
		t.Fatalf("Code = %s, want %s", conflict.Code, LifecycleErrorCodeStateConflict)
	}
	if conflict.Stage != LifecycleStageStateTransition {
		t.Fatalf("Stage = %s, want %s", conflict.Stage, LifecycleStageStateTransition)
	}
	if conflict.FieldPath != "state" {
		t.Fatalf("FieldPath = %q, want state", conflict.FieldPath)
	}
	assertNoSensitiveFragments(t, conflict.SafeMessage)
	if strings.Contains(conflict.SafeMessage, "COMPLETED") || strings.Contains(conflict.SafeMessage, "RUNNING") {
		t.Fatalf("state conflict safe message leaked raw state values: %q", conflict.SafeMessage)
	}
}

func TestMapDownstreamFailureSuppressesRawErrorPayload(t *testing.T) {
	raw := errors.New("password=hunter2 host=db.internal:5432 SELECT * FROM users generated data: [{email:'a@example.test'}]")

	mapped := MapDownstreamFailure(LifecycleStageGeneration, "generation", raw)

	if mapped.Code != LifecycleErrorCodeDownstreamFailure {
		t.Fatalf("Code = %s, want %s", mapped.Code, LifecycleErrorCodeDownstreamFailure)
	}
	if mapped.Stage != LifecycleStageGeneration {
		t.Fatalf("Stage = %s, want %s", mapped.Stage, LifecycleStageGeneration)
	}
	if mapped.FieldPath != "generation" {
		t.Fatalf("FieldPath = %q, want generation", mapped.FieldPath)
	}
	assertNoSensitiveFragments(t, mapped.SafeMessage)
	if strings.Contains(mapped.SafeMessage, raw.Error()) {
		t.Fatalf("downstream safe message leaked raw error payload: %q", mapped.SafeMessage)
	}
}

func TestNewLifecycleErrorReplacesSensitiveSafeMessage(t *testing.T) {
	mapped := NewLifecycleError(
		LifecycleErrorCodeInvalidReference,
		LifecycleStagePrecheck,
		"connection",
		"connection password secret leaked with user SQL SELECT * FROM users",
	)

	if mapped.Code != LifecycleErrorCodeInvalidReference {
		t.Fatalf("Code = %s, want %s", mapped.Code, LifecycleErrorCodeInvalidReference)
	}
	if mapped.Stage != LifecycleStagePrecheck {
		t.Fatalf("Stage = %s, want %s", mapped.Stage, LifecycleStagePrecheck)
	}
	if mapped.FieldPath != "connection" {
		t.Fatalf("FieldPath = %q, want connection", mapped.FieldPath)
	}
	assertNoSensitiveFragments(t, mapped.SafeMessage)
	if mapped.SafeMessage == "connection password secret leaked with user SQL SELECT * FROM users" {
		t.Fatal("expected sensitive safe message candidate to be replaced")
	}
}

func TestNewLifecycleErrorSanitizesPublicSummaryFields(t *testing.T) {
	mapped := NewLifecycleError(
		"RAW_DRIVER_CODE password=hunter2",
		"PRECHECK host=db.internal SELECT * FROM users",
		"generation.generatedData password=secret",
		"generated data: [{email:'a@example.test'}] connection string password=secret",
	)

	if mapped.Code != LifecycleErrorCodeSensitiveValueNotAllowed {
		t.Fatalf("Code = %s, want %s", mapped.Code, LifecycleErrorCodeSensitiveValueNotAllowed)
	}
	if mapped.Stage != LifecycleStagePrecheck {
		t.Fatalf("Stage = %s, want %s", mapped.Stage, LifecycleStagePrecheck)
	}
	if mapped.FieldPath != "lifecycle.error" {
		t.Fatalf("FieldPath = %q, want lifecycle.error", mapped.FieldPath)
	}
	assertNoSensitiveLifecycleErrorContent(t, mapped)
}

func assertLifecycleError(t *testing.T, issues []LifecycleError, stage LifecycleStage, fieldPath string, code LifecycleErrorCode) {
	t.Helper()
	for _, issue := range issues {
		if issue.Stage == stage && issue.FieldPath == fieldPath && issue.Code == code {
			assertNoSensitiveFragments(t, issue.SafeMessage)
			return
		}
	}
	t.Fatalf("expected lifecycle error stage=%s fieldPath=%s code=%s in %#v", stage, fieldPath, code, issues)
}

func assertNoSensitiveLifecycleErrorContent(t *testing.T, issue LifecycleError) {
	t.Helper()
	assertNoSensitiveFragments(t, issue.Code.String())
	assertNoSensitiveFragments(t, issue.Stage.String())
	assertNoSensitiveFragments(t, issue.FieldPath)
	assertNoSensitiveFragments(t, issue.SafeMessage)
}

func assertNoSensitiveFragments(t *testing.T, value string) {
	t.Helper()
	lower := strings.ToLower(value)
	for _, fragment := range []string{"password", "hunter2", "select *", "host=", "db.internal", "generated data", "a@example.test", "secret", "user sql"} {
		if strings.Contains(lower, fragment) {
			t.Fatalf("safe message %q contains sensitive fragment %q", value, fragment)
		}
	}
	if strings.TrimSpace(value) == "" {
		t.Fatal("safe message must not be blank")
	}
}
