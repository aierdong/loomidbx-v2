package plan

import (
	"reflect"
	"strings"
	"testing"
)

func TestPlanIssuePublicShapeContainsOnlySafeSummaryFields(t *testing.T) {
	issueType := reflect.TypeFor[PlanIssue]()
	wantFields := []string{"Code", "Stage", "FieldPath", "SafeMessage", "Blocking"}
	if issueType.NumField() != len(wantFields) {
		t.Fatalf("PlanIssue has %d public fields, want only safe summary fields %v", issueType.NumField(), wantFields)
	}
	for index, want := range wantFields {
		field := issueType.Field(index)
		if field.Name != want {
			t.Fatalf("PlanIssue field[%d] = %s, want %s", index, field.Name, want)
		}
	}
}

func TestNewPlanIssueReplacesSensitiveSafeMessage(t *testing.T) {
	issue := NewPlanIssue(
		PlanErrorCodeInvalidReference,
		PlanStageRelationMapping,
		"relations[0].source",
		"connection password secret leaked with user SQL SELECT * FROM users generated data",
		true,
	)

	if issue.Code != PlanErrorCodeInvalidReference {
		t.Fatalf("Code = %s, want %s", issue.Code, PlanErrorCodeInvalidReference)
	}
	if issue.Stage != PlanStageRelationMapping {
		t.Fatalf("Stage = %s, want %s", issue.Stage, PlanStageRelationMapping)
	}
	if issue.FieldPath != "relations[0].source" {
		t.Fatalf("FieldPath = %q, want relations[0].source", issue.FieldPath)
	}
	if !issue.Blocking {
		t.Fatal("expected blocking issue")
	}
	assertNoSensitivePlanFragments(t, issue.SafeMessage)
	if strings.Contains(issue.SafeMessage, "SELECT") || strings.Contains(issue.SafeMessage, "password") {
		t.Fatalf("safe message leaked sensitive content: %q", issue.SafeMessage)
	}
}

func TestNewPlanIssueReplacesConnectionStringSafeMessage(t *testing.T) {
	issue := NewPlanIssue(
		PlanErrorCodeInvalidReference,
		PlanStageRelationMapping,
		"relations[0].source",
		"postgres://app:pass123@db.internal:5432/tenant",
		true,
	)

	if issue.Code != PlanErrorCodeInvalidReference {
		t.Fatalf("Code = %s, want %s", issue.Code, PlanErrorCodeInvalidReference)
	}
	if issue.SafeMessage != defaultPlanSafeMessage(PlanErrorCodeInvalidReference) {
		t.Fatalf("SafeMessage = %q, want default invalid reference message", issue.SafeMessage)
	}
	assertNoSensitivePlanIssueContent(t, issue)
}

func TestNewPlanIssueSanitizesPublicSummaryFields(t *testing.T) {
	issue := NewPlanIssue(
		"RAW_DRIVER_CODE password=hunter2",
		"SORT host=db.internal SELECT * FROM users",
		"graph.generatedData password=secret",
		"generated data: [{email:'a@example.test'}] connection string password=secret",
		true,
	)

	if issue.Code != PlanErrorCodeSensitiveValueNotAllowed {
		t.Fatalf("Code = %s, want %s", issue.Code, PlanErrorCodeSensitiveValueNotAllowed)
	}
	if issue.Stage != PlanStagePrecheck {
		t.Fatalf("Stage = %s, want %s", issue.Stage, PlanStagePrecheck)
	}
	if issue.FieldPath != "plan.issue" {
		t.Fatalf("FieldPath = %q, want plan.issue", issue.FieldPath)
	}
	assertNoSensitivePlanIssueContent(t, issue)
}

func TestPlanPrecheckAddBlockingIssueCreatesSafeIssue(t *testing.T) {
	result := PlanPrecheckResult{Passed: true}

	result.addBlockingIssue("graph.nodes[0].id", PlanErrorCodeRequired, PlanStageGraphBuild, "node identity is required")

	if result.Passed {
		t.Fatal("expected blocking issue to fail precheck")
	}
	if len(result.BlockingErrors) != 1 {
		t.Fatalf("BlockingErrors length = %d, want 1", len(result.BlockingErrors))
	}
	issue := result.BlockingErrors[0]
	if issue.Code != PlanErrorCodeRequired || issue.Stage != PlanStageGraphBuild || issue.FieldPath != "graph.nodes[0].id" || !issue.Blocking {
		t.Fatalf("unexpected blocking issue: %#v", issue)
	}
	assertNoSensitivePlanFragments(t, issue.SafeMessage)
}

func assertNoSensitivePlanIssueContent(t *testing.T, issue PlanIssue) {
	t.Helper()
	assertNoSensitivePlanFragments(t, issue.Code.String())
	assertNoSensitivePlanFragments(t, issue.Stage.String())
	assertNoSensitivePlanFragments(t, issue.FieldPath)
	assertNoSensitivePlanFragments(t, issue.SafeMessage)
}

func assertNoSensitivePlanFragments(t *testing.T, value string) {
	t.Helper()
	lower := strings.ToLower(value)
	for _, fragment := range []string{"password", "hunter2", "pass123", "postgres://", "select *", "host=", "db.internal", "generated data", "a@example.test", "secret", "user sql"} {
		if strings.Contains(lower, fragment) {
			t.Fatalf("safe plan value %q contains sensitive fragment %q", value, fragment)
		}
	}
	if strings.TrimSpace(value) == "" {
		t.Fatal("safe plan value must not be blank")
	}
}
