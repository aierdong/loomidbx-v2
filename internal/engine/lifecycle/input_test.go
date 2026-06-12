package lifecycle

import (
	"testing"
	"time"

	domainexecution "github.com/gerdong/loomidbx/internal/domain/execution"
)

func TestMapExecutionInputFromGenerationJobAcceptsValidSnapshot(t *testing.T) {
	job := validGenerationJobSnapshot()

	input, precheck := MapExecutionInputFromGenerationJob(job)

	if !precheck.Passed {
		t.Fatalf("expected valid snapshot to pass precheck, got %#v", precheck.BlockingErrors)
	}
	if input == nil {
		t.Fatal("expected runnable execution input")
	}
	if input.TaskID != job.Task.ID {
		t.Fatalf("TaskID = %d, want %d", input.TaskID, job.Task.ID)
	}
	if input.ProjectID != job.Task.ProjectID {
		t.Fatalf("ProjectID = %d, want %d", input.ProjectID, job.Task.ProjectID)
	}
	if input.TaskName != job.Task.TaskName {
		t.Fatalf("TaskName = %q, want %q", input.TaskName, job.Task.TaskName)
	}
	if len(input.Tables) != len(job.TableResults) {
		t.Fatalf("Tables length = %d, want %d", len(input.Tables), len(job.TableResults))
	}
	if input.Tables[0].TableResultID != job.TableResults[0].ID {
		t.Fatalf("TableResultID = %d, want %d", input.Tables[0].TableResultID, job.TableResults[0].ID)
	}
	if input.Tables[0].ExecutionTaskID != job.TableResults[0].ExecutionTaskID {
		t.Fatalf("ExecutionTaskID = %d, want %d", input.Tables[0].ExecutionTaskID, job.TableResults[0].ExecutionTaskID)
	}
	if input.Tables[0].TableID == nil || *input.Tables[0].TableID != *job.TableResults[0].TableID {
		t.Fatalf("TableID = %v, want %v", input.Tables[0].TableID, job.TableResults[0].TableID)
	}
	if input.Tables[0].SchemaName != job.TableResults[0].SchemaNameSnapshot {
		t.Fatalf("SchemaName = %q, want %q", input.Tables[0].SchemaName, job.TableResults[0].SchemaNameSnapshot)
	}
	if input.Tables[0].TableName != job.TableResults[0].TableNameSnapshot {
		t.Fatalf("TableName = %q, want %q", input.Tables[0].TableName, job.TableResults[0].TableNameSnapshot)
	}
	if input.Tables[0].ExecutionOrder != job.TableResults[0].ExecutionOrder {
		t.Fatalf("ExecutionOrder = %d, want %d", input.Tables[0].ExecutionOrder, job.TableResults[0].ExecutionOrder)
	}
}

func TestMapExecutionInputFromGenerationJobRejectsMissingInputBoundary(t *testing.T) {
	job := validGenerationJobSnapshot()
	job.Task.ID = 0
	job.Task.ProjectID = 0
	job.Task.TaskName = "   "
	job.TableResults = nil

	input, precheck := MapExecutionInputFromGenerationJob(job)

	if input != nil {
		t.Fatalf("expected no runnable execution input, got %#v", input)
	}
	if precheck.Passed {
		t.Fatal("expected missing boundary fields to fail precheck")
	}
	assertBlockingIssue(t, precheck, "task.id", PrecheckIssueCodeRequired)
	assertBlockingIssue(t, precheck, "task.projectId", PrecheckIssueCodeInvalidReference)
	assertBlockingIssue(t, precheck, "task.taskName", PrecheckIssueCodeRequired)
	assertBlockingIssue(t, precheck, "tableResults", PrecheckIssueCodeRequired)
}

func TestMapExecutionInputFromGenerationJobRejectsInvalidTableBoundary(t *testing.T) {
	job := validGenerationJobSnapshot()
	job.TableResults[0].ExecutionTaskID = 999
	job.TableResults[0].TableNameSnapshot = ""
	job.TableResults[0].SchemaNameSnapshot = " "
	job.TableResults[0].ExecutionOrder = 0

	input, precheck := MapExecutionInputFromGenerationJob(job)

	if input != nil {
		t.Fatalf("expected no runnable execution input, got %#v", input)
	}
	if precheck.Passed {
		t.Fatal("expected invalid table boundary to fail precheck")
	}
	assertBlockingIssue(t, precheck, "tableResults[0].executionTaskId", PrecheckIssueCodeInvalidReference)
	assertBlockingIssue(t, precheck, "tableResults[0].tableNameSnapshot", PrecheckIssueCodeRequired)
	assertBlockingIssue(t, precheck, "tableResults[0].schemaNameSnapshot", PrecheckIssueCodeRequired)
	assertBlockingIssue(t, precheck, "tableResults[0].executionOrder", PrecheckIssueCodeInvalidRange)
}

func TestExecutionInputDoesNotExposeRuntimeOrDatabaseDependencies(t *testing.T) {
	job := validGenerationJobSnapshot()

	input, precheck := MapExecutionInputFromGenerationJob(job)

	if !precheck.Passed || input == nil {
		t.Fatalf("expected valid snapshot to map, input=%#v precheck=%#v", input, precheck)
	}
	type executionInputShape struct {
		TaskID    int64
		ProjectID int64
		TaskName  string
		Tables    []ExecutionTableInput
	}
	shape := executionInputShape(*input)
	if shape.TaskID == 0 || shape.ProjectID == 0 || shape.TaskName == "" || len(shape.Tables) == 0 {
		t.Fatalf("expected minimal execution input shape to be populated, got %#v", shape)
	}
}

func assertBlockingIssue(t *testing.T, result PrecheckResult, fieldPath string, code PrecheckIssueCode) {
	t.Helper()
	for _, issue := range result.BlockingErrors {
		if issue.FieldPath == fieldPath && issue.Code == code {
			return
		}
	}
	t.Fatalf("expected blocking issue %s/%s in %#v", fieldPath, code, result.BlockingErrors)
}

func validGenerationJobSnapshot() domainexecution.GenerationJob {
	now := time.Date(2026, 6, 12, 9, 0, 0, 0, time.UTC)
	tableID := int64(301)
	return domainexecution.GenerationJob{
		Task: domainexecution.ExecutionTask{
			ID:        101,
			ProjectID: 201,
			TaskName:  "nightly seed",
			Status:    domainexecution.ExecutionTaskStatusRunning,
			StartedAt: now,
			CreatedAt: now,
		},
		TableResults: []domainexecution.ExecutionTableResult{
			{
				ID:                 401,
				ExecutionTaskID:    101,
				TableID:            &tableID,
				SchemaNameSnapshot: "public",
				TableNameSnapshot:  "users",
				RowsWritten:        0,
				Status:             domainexecution.ExecutionTableStatusPending,
				ExecutionOrder:     1,
				CreatedAt:          now,
				UpdatedAt:          now,
			},
		},
	}
}
