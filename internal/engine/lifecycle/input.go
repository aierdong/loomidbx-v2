// Package lifecycle contains engine-only execution lifecycle input and precheck models.
package lifecycle

import (
	"strconv"
	"strings"

	domainexecution "github.com/gerdong/loomidbx/internal/domain/execution"
)

// PrecheckIssueCode identifies a stable machine-readable lifecycle precheck issue category.
type PrecheckIssueCode = LifecycleErrorCode

const (
	// PrecheckIssueCodeRequired reports that a required lifecycle input field is missing or blank.
	PrecheckIssueCodeRequired PrecheckIssueCode = LifecycleErrorCodeRequired

	// PrecheckIssueCodeInvalidReference reports that a lifecycle input identity or parent reference is invalid.
	PrecheckIssueCodeInvalidReference PrecheckIssueCode = LifecycleErrorCodeInvalidReference

	// PrecheckIssueCodeInvalidRange reports that a lifecycle input numeric boundary is outside the accepted range.
	PrecheckIssueCodeInvalidRange PrecheckIssueCode = LifecycleErrorCodeInvalidRange
)

// PrecheckIssue is a field-level lifecycle precheck problem with a safe diagnostic message.
type PrecheckIssue = LifecycleError

// PrecheckResult summarizes whether lifecycle precheck allows execution to start.
type PrecheckResult struct {
	// Passed reports whether no blocking errors were found.
	Passed bool

	// BlockingErrors contains field-level issues that prevent a runnable execution input from being created.
	BlockingErrors []PrecheckIssue

	// Warnings contains non-blocking precheck issues. The input mapper currently emits no warnings.
	Warnings []PrecheckIssue
}

// ExecutionInput is the engine-only value object produced from an execution task snapshot.
type ExecutionInput struct {
	// TaskID stores the execution task identity used by the lifecycle run.
	TaskID int64

	// ProjectID stores the owning project reference for this execution.
	ProjectID int64

	// TaskName stores the task name snapshot used for downstream lifecycle seams.
	TaskName string

	// Tables stores table-level execution boundaries without dependency, row planning, batch, or result artifacts.
	Tables []ExecutionTableInput
}

// ExecutionTableInput is the minimum table-level boundary needed by later engine seams.
type ExecutionTableInput struct {
	// TableResultID stores the table result record identity when the snapshot already has one.
	TableResultID int64

	// ExecutionTaskID stores the parent execution task identity captured on the table result.
	ExecutionTaskID int64

	// TableID stores the optional schema table identity while allowing historical snapshots without a live table reference.
	TableID *int64

	// SchemaName stores the schema name snapshot captured for the table boundary.
	SchemaName string

	// TableName stores the table name snapshot captured for the table boundary.
	TableName string

	// ExecutionOrder stores the existing table result order boundary without recalculating dependencies.
	ExecutionOrder int
}

// MapExecutionInputFromGenerationJob validates a generation job snapshot and maps it into an engine-only execution input.
func MapExecutionInputFromGenerationJob(job domainexecution.GenerationJob) (*ExecutionInput, PrecheckResult) {
	result := PrecheckResult{Passed: true}
	validateTaskBoundary(&result, job.Task)
	validateTableBoundary(&result, job.Task.ID, job.TableResults)
	if len(result.BlockingErrors) > 0 {
		result.Passed = false
		return nil, result
	}

	input := &ExecutionInput{
		TaskID:    job.Task.ID,
		ProjectID: job.Task.ProjectID,
		TaskName:  strings.TrimSpace(job.Task.TaskName),
		Tables:    make([]ExecutionTableInput, 0, len(job.TableResults)),
	}
	for _, tableResult := range job.TableResults {
		input.Tables = append(input.Tables, ExecutionTableInput{
			TableResultID:   tableResult.ID,
			ExecutionTaskID: tableResult.ExecutionTaskID,
			TableID:         cloneInt64Pointer(tableResult.TableID),
			SchemaName:      strings.TrimSpace(tableResult.SchemaNameSnapshot),
			TableName:       strings.TrimSpace(tableResult.TableNameSnapshot),
			ExecutionOrder:  tableResult.ExecutionOrder,
		})
	}

	return input, result
}

func validateTaskBoundary(result *PrecheckResult, task domainexecution.ExecutionTask) {
	if task.ID <= 0 {
		result.addBlockingIssue("task.id", PrecheckIssueCodeRequired, "task identity is required before execution can start")
	}
	if task.ProjectID <= 0 {
		result.addBlockingIssue("task.projectId", PrecheckIssueCodeInvalidReference, "project reference is required before execution can start")
	}
	if strings.TrimSpace(task.TaskName) == "" {
		result.addBlockingIssue("task.taskName", PrecheckIssueCodeRequired, "task name is required before execution can start")
	}
}

func validateTableBoundary(result *PrecheckResult, taskID int64, tableResults []domainexecution.ExecutionTableResult) {
	if len(tableResults) == 0 {
		result.addBlockingIssue("tableResults", PrecheckIssueCodeRequired, "at least one table boundary is required before execution can start")
		return
	}

	for index, tableResult := range tableResults {
		prefix := "tableResults[" + strconv.Itoa(index) + "]"
		if tableResult.ExecutionTaskID <= 0 {
			result.addBlockingIssue(prefix+".executionTaskId", PrecheckIssueCodeInvalidReference, "table boundary must reference its execution task")
		} else if taskID > 0 && tableResult.ExecutionTaskID != taskID {
			result.addBlockingIssue(prefix+".executionTaskId", PrecheckIssueCodeInvalidReference, "table boundary must reference the submitted execution task")
		}
		if strings.TrimSpace(tableResult.TableNameSnapshot) == "" {
			result.addBlockingIssue(prefix+".tableNameSnapshot", PrecheckIssueCodeRequired, "table name snapshot is required before execution can start")
		}
		if strings.TrimSpace(tableResult.SchemaNameSnapshot) == "" {
			result.addBlockingIssue(prefix+".schemaNameSnapshot", PrecheckIssueCodeRequired, "schema name snapshot is required before execution can start")
		}
		if tableResult.ExecutionOrder < 1 {
			result.addBlockingIssue(prefix+".executionOrder", PrecheckIssueCodeInvalidRange, "table execution order boundary must be greater than or equal to one")
		}
		if tableResult.TableID != nil && *tableResult.TableID <= 0 {
			result.addBlockingIssue(prefix+".tableId", PrecheckIssueCodeInvalidReference, "table reference must be positive when present")
		}
	}
}

func (r *PrecheckResult) addBlockingIssue(fieldPath string, code PrecheckIssueCode, safeMessage string) {
	r.BlockingErrors = append(r.BlockingErrors, NewLifecycleError(code, LifecycleStageInputValidation, fieldPath, safeMessage))
}

func cloneInt64Pointer(value *int64) *int64 {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}
