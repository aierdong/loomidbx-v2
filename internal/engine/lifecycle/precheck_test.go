package lifecycle

import (
	"testing"
	"time"
)

func TestAggregatePrecheckMergesInputStateAndDownstreamIssues(t *testing.T) {
	job := validGenerationJobSnapshot()
	input, base := MapExecutionInputFromGenerationJob(job)
	if input == nil || !base.Passed {
		t.Fatalf("test fixture must map valid input, input=%#v base=%#v", input, base)
	}
	base.Warnings = append(base.Warnings, NewLifecycleError(
		LifecycleErrorCodeInvalidRange,
		LifecycleStageInputValidation,
		"task.batchSize",
		"optional execution batch size will use the engine default",
	))
	downstream := PrecheckResult{
		Passed: false,
		BlockingErrors: []PrecheckIssue{
			NewLifecycleError(LifecycleErrorCodeInvalidReference, LifecycleStagePrecheck, "planner.schema", "planner prerequisite is not available"),
		},
		Warnings: []PrecheckIssue{
			NewLifecycleError(LifecycleErrorCodeInvalidRange, LifecycleStagePrecheck, "result.summary", "result summary may be partial"),
		},
	}

	result := AggregatePrecheck(input, NewLifecycle(), base, downstream)

	if result.Passed {
		t.Fatal("expected downstream blocking error to fail aggregate precheck")
	}
	assertBlockingIssue(t, result, "planner.schema", LifecycleErrorCodeInvalidReference)
	assertWarningIssue(t, result, "task.batchSize", LifecycleErrorCodeInvalidRange)
	assertWarningIssue(t, result, "result.summary", LifecycleErrorCodeInvalidRange)
}

func TestAggregatePrecheckAllowsWarningsOnly(t *testing.T) {
	job := validGenerationJobSnapshot()
	input, base := MapExecutionInputFromGenerationJob(job)
	if input == nil || !base.Passed {
		t.Fatalf("test fixture must map valid input, input=%#v base=%#v", input, base)
	}
	downstream := PrecheckResult{
		Passed: true,
		Warnings: []PrecheckIssue{
			NewLifecycleError(LifecycleErrorCodeInvalidRange, LifecycleStagePrecheck, "planner.optional", "optional planner hint is not configured"),
		},
	}

	result := AggregatePrecheck(input, NewLifecycle(), base, downstream)

	if !result.Passed {
		t.Fatalf("warnings-only precheck should pass, blocking=%#v", result.BlockingErrors)
	}
	if len(result.BlockingErrors) != 0 {
		t.Fatalf("warnings-only precheck returned blocking errors: %#v", result.BlockingErrors)
	}
	assertWarningIssue(t, result, "planner.optional", LifecycleErrorCodeInvalidRange)
}

func TestAggregatePrecheckBlocksNonPrecheckableLifecycleWithoutRunning(t *testing.T) {
	job := validGenerationJobSnapshot()
	input, base := MapExecutionInputFromGenerationJob(job)
	if input == nil || !base.Passed {
		t.Fatalf("test fixture must map valid input, input=%#v base=%#v", input, base)
	}
	machine := NewLifecycle()
	startedAt := time.Date(2026, 6, 12, 11, 0, 0, 0, time.UTC)
	for index, state := range []LifecycleState{LifecycleStatePrechecking, LifecycleStateReady} {
		transition := machine.TransitionTo(state, startedAt.Add(time.Duration(index)*time.Minute))
		if !transition.Accepted {
			t.Fatalf("fixture transition to %s rejected: %#v", state, transition.Error)
		}
	}

	result := AggregatePrecheck(input, machine, base)

	if result.Passed {
		t.Fatal("expected non-precheckable lifecycle state to block startup")
	}
	assertBlockingIssue(t, result, "state", LifecycleErrorCodeStateConflict)
	if machine.State() == LifecycleStateRunning {
		t.Fatal("precheck aggregator must not move lifecycle into RUNNING")
	}
	if machine.State() != LifecycleStateReady {
		t.Fatalf("precheck aggregator changed lifecycle state to %s, want %s", machine.State(), LifecycleStateReady)
	}
}

func TestAggregatePrecheckBlocksMissingExecutionInputEvenWhenBaseResultIsEmpty(t *testing.T) {
	result := AggregatePrecheck(nil, NewLifecycle(), PrecheckResult{Passed: true})

	if result.Passed {
		t.Fatal("expected missing execution input to fail aggregate precheck")
	}
	assertBlockingIssue(t, result, "input", LifecycleErrorCodeRequired)
}

func assertWarningIssue(t *testing.T, result PrecheckResult, fieldPath string, code PrecheckIssueCode) {
	t.Helper()
	for _, issue := range result.Warnings {
		if issue.FieldPath == fieldPath && issue.Code == code {
			return
		}
	}
	t.Fatalf("expected warning issue %s/%s in %#v", fieldPath, code, result.Warnings)
}
