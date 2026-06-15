package plan

import (
	"testing"

	domainproject "github.com/gerdong/loomidbx/internal/domain/project"
)

func TestBuildPlanInputMapsValidProjectTablesToNodeInputs(t *testing.T) {
	projectSnapshot := domainproject.Project{ID: 101, Name: "demo"}
	tables := []domainproject.ProjectTable{
		{ID: 301, ProjectID: 101, TableID: 201, ExecutionOrder: 2},
		{ID: 302, ProjectID: 101, TableID: 202, ExecutionOrder: 1},
	}

	input, result := BuildPlanInput(projectSnapshot, tables, nil, nil, nil)

	if !result.Passed {
		t.Fatalf("expected valid ProjectTable snapshots to pass, got %#v", result.BlockingErrors)
	}
	if input == nil {
		t.Fatal("expected plan input")
	}
	if input.ProjectID != projectSnapshot.ID {
		t.Fatalf("ProjectID = %d, want %d", input.ProjectID, projectSnapshot.ID)
	}
	if len(input.Tables) != len(tables) {
		t.Fatalf("Tables length = %d, want %d", len(input.Tables), len(tables))
	}
	assertPlanTableInput(t, input.Tables[0], PlanTableInput{ProjectTableID: 301, ProjectID: 101, TableID: 201, ExistingExecutionOrder: 2})
	assertPlanTableInput(t, input.Tables[1], PlanTableInput{ProjectTableID: 302, ProjectID: 101, TableID: 202, ExistingExecutionOrder: 1})
}

func TestBuildPlanInputRejectsMissingProjectTableNodeIdentity(t *testing.T) {
	projectSnapshot := domainproject.Project{ID: 101}
	tables := []domainproject.ProjectTable{
		{ID: 0, ProjectID: 101, TableID: 201, ExecutionOrder: 1},
		{ID: 302, ProjectID: 0, TableID: 202, ExecutionOrder: 2},
		{ID: 303, ProjectID: 101, TableID: 0, ExecutionOrder: 3},
	}

	input, result := BuildPlanInput(projectSnapshot, tables, nil, nil, nil)

	if input != nil {
		t.Fatalf("expected no plan input for invalid node boundary, got %#v", input)
	}
	if result.Passed {
		t.Fatal("expected missing node identities to fail precheck")
	}
	assertPlanIssue(t, result.BlockingErrors, "tables[0].id", PlanErrorCodeRequired, PlanStageInputMapping)
	assertPlanIssue(t, result.BlockingErrors, "tables[1].projectId", PlanErrorCodeInvalidReference, PlanStageInputMapping)
	assertPlanIssue(t, result.BlockingErrors, "tables[2].tableId", PlanErrorCodeInvalidReference, PlanStageInputMapping)
}

func TestBuildPlanInputRejectsDuplicateSchemaTableNodes(t *testing.T) {
	projectSnapshot := domainproject.Project{ID: 101}
	tables := []domainproject.ProjectTable{
		{ID: 301, ProjectID: 101, TableID: 201, ExecutionOrder: 1},
		{ID: 302, ProjectID: 101, TableID: 201, ExecutionOrder: 2},
	}

	input, result := BuildPlanInput(projectSnapshot, tables, nil, nil, nil)

	if input != nil {
		t.Fatalf("expected duplicate table IDs to block plan input, got %#v", input)
	}
	if result.Passed {
		t.Fatal("expected duplicate table node to fail precheck")
	}
	assertPlanIssue(t, result.BlockingErrors, "tables[1].tableId", PlanErrorCodeDuplicateNode, PlanStageInputMapping)
}

func assertPlanTableInput(t *testing.T, got PlanTableInput, want PlanTableInput) {
	t.Helper()
	if got != want {
		t.Fatalf("PlanTableInput = %#v, want %#v", got, want)
	}
}

func assertPlanIssue(t *testing.T, issues []PlanIssue, fieldPath string, code PlanErrorCode, stage PlanStage) {
	t.Helper()
	for _, issue := range issues {
		if issue.FieldPath == fieldPath && issue.Code == code && issue.Stage == stage && issue.Blocking {
			if issue.SafeMessage == "" {
				t.Fatalf("issue has blank safe message: %#v", issue)
			}
			return
		}
	}
	t.Fatalf("expected blocking issue fieldPath=%s code=%s stage=%s in %#v", fieldPath, code, stage, issues)
}
