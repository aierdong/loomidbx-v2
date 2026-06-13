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

func TestAggregatePrecheckSanitizesMalformedDownstreamIssues(t *testing.T) {
	job := validGenerationJobSnapshot()
	input, base := MapExecutionInputFromGenerationJob(job)
	if input == nil || !base.Passed {
		t.Fatalf("test fixture must map valid input, input=%#v base=%#v", input, base)
	}
	downstream := PrecheckResult{
		Passed: false,
		BlockingErrors: []PrecheckIssue{
			{
				Code:        "RAW_DRIVER_CODE password=hunter2",
				Stage:       LifecycleStagePrecheck,
				FieldPath:   "planner.sql SELECT * FROM users host=db.internal",
				SafeMessage: "password=hunter2 host=db.internal:5432 SELECT * FROM users generated data: [{email:'a@example.test'}]",
			},
		},
		Warnings: []PrecheckIssue{
			{
				Code:        LifecycleErrorCodeInvalidRange,
				Stage:       "PRECHECK password=hunter2 SELECT * FROM users",
				FieldPath:   "result.summary",
				SafeMessage: "generated data contains token=raw-token and connection string password=secret",
			},
		},
	}

	result := AggregatePrecheck(input, NewLifecycle(), base, downstream)

	if result.Passed {
		t.Fatal("malformed blocking downstream issue should fail aggregate precheck")
	}
	if len(result.BlockingErrors) != 1 {
		t.Fatalf("BlockingErrors length = %d, want 1", len(result.BlockingErrors))
	}
	blocking := result.BlockingErrors[0]
	if blocking.Code != LifecycleErrorCodeSensitiveValueNotAllowed {
		t.Fatalf("BlockingErrors[0].Code = %s, want %s", blocking.Code, LifecycleErrorCodeSensitiveValueNotAllowed)
	}
	if blocking.FieldPath != "precheck.issue" {
		t.Fatalf("BlockingErrors[0].FieldPath = %q, want precheck.issue", blocking.FieldPath)
	}
	assertNoSensitiveLifecycleErrorContent(t, blocking)
	if len(result.Warnings) != 1 {
		t.Fatalf("Warnings length = %d, want 1", len(result.Warnings))
	}
	warning := result.Warnings[0]
	if warning.Code != LifecycleErrorCodeSensitiveValueNotAllowed {
		t.Fatalf("Warnings[0].Code = %s, want %s", warning.Code, LifecycleErrorCodeSensitiveValueNotAllowed)
	}
	if warning.Stage != LifecycleStagePrecheck {
		t.Fatalf("Warnings[0].Stage = %s, want %s", warning.Stage, LifecycleStagePrecheck)
	}
	if warning.FieldPath != "precheck.issue" {
		t.Fatalf("Warnings[0].FieldPath = %q, want precheck.issue", warning.FieldPath)
	}
	assertNoSensitiveLifecycleErrorContent(t, warning)
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

func TestCoordinatorStartsWhenAggregatePrecheckHasWarningsOnly(t *testing.T) {
	job := validGenerationJobSnapshot()
	warning := NewLifecycleError(
		LifecycleErrorCodeInvalidRange,
		LifecycleStagePrecheck,
		"planner.optional",
		"optional planner hint is not configured",
	)
	coordinator := NewLifecycleCoordinatorWithPorts(
		sequenceClock(
			time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC),
			time.Date(2026, 6, 12, 12, 1, 0, 0, time.UTC),
			time.Date(2026, 6, 12, 12, 2, 0, 0, time.UTC),
			time.Date(2026, 6, 12, 12, 3, 0, 0, time.UTC),
		),
		DownstreamPorts{Precheck: stubPrechecker{result: PrecheckResult{Passed: true, Warnings: []PrecheckIssue{warning}}}},
	)

	result := coordinator.Run(job)

	if !result.Precheck.Passed {
		t.Fatalf("warnings-only precheck should allow startup, blocking=%#v", result.Precheck.BlockingErrors)
	}
	assertWarningIssue(t, result.Precheck, "planner.optional", LifecycleErrorCodeInvalidRange)
	if result.StartedAt == nil {
		t.Fatal("expected lifecycle to start when aggregate precheck only has warnings")
	}
	if result.State != LifecycleStateCompleted {
		t.Fatalf("State = %s, want %s", result.State, LifecycleStateCompleted)
	}
	assertTransitionPath(t, result.Transitions, []LifecycleState{
		LifecycleStateInitialized,
		LifecycleStatePrechecking,
		LifecycleStateReady,
		LifecycleStateRunning,
		LifecycleStateCompleted,
	})
}

func TestCoordinatorDoesNotStartWhenAggregatePrecheckHasBlockingError(t *testing.T) {
	job := validGenerationJobSnapshot()
	blocking := NewLifecycleError(
		LifecycleErrorCodeInvalidReference,
		LifecycleStagePrecheck,
		"planner.schema",
		"planner prerequisite is not available",
	)
	coordinator := NewLifecycleCoordinatorWithPorts(
		sequenceClock(
			time.Date(2026, 6, 12, 13, 0, 0, 0, time.UTC),
			time.Date(2026, 6, 12, 13, 1, 0, 0, time.UTC),
		),
		DownstreamPorts{Precheck: stubPrechecker{result: PrecheckResult{Passed: false, BlockingErrors: []PrecheckIssue{blocking}}}},
	)

	result := coordinator.Run(job)

	if result.Precheck.Passed {
		t.Fatal("blocking precheck error should prevent startup")
	}
	assertBlockingIssue(t, result.Precheck, "planner.schema", LifecycleErrorCodeInvalidReference)
	if result.StartedAt != nil {
		t.Fatalf("expected lifecycle not to start, StartedAt=%s", result.StartedAt.Format(time.RFC3339Nano))
	}
	if result.State != LifecycleStateFailed {
		t.Fatalf("State = %s, want %s", result.State, LifecycleStateFailed)
	}
	assertTransitionPath(t, result.Transitions, []LifecycleState{
		LifecycleStateInitialized,
		LifecycleStatePrechecking,
		LifecycleStateFailed,
	})
	if result.Failure == nil || result.Failure.FieldPath != "planner.schema" {
		t.Fatalf("expected precheck failure summary for planner.schema, got %#v", result.Failure)
	}
}

type stubPrechecker struct {
	result PrecheckResult
}

func (s stubPrechecker) Precheck(context DownstreamContext) PrecheckResult {
	return s.result
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
