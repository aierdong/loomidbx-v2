package lifecycle

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCoordinatorInvokesReplaceableDownstreamPortsInOrder(t *testing.T) {
	job := validGenerationJobSnapshot()
	var calls []string
	ports := DownstreamPorts{
		Precheck:   fakePrecheckPort{t: t, calls: &calls, name: "precheck"},
		Planner:    fakePlannerPort{t: t, calls: &calls, name: "planner", result: NewDownstreamStageSuccess("opaque-plan")},
		Generation: fakeGenerationPort{t: t, calls: &calls, name: "generation", wantPlanArtifact: "opaque-plan", result: NewDownstreamStageSuccess("opaque-generation")},
		Result:     fakeResultPort{t: t, calls: &calls, name: "result", wantPlanArtifact: "opaque-plan", wantGenerationArtifact: "opaque-generation", result: NewDownstreamStageSuccess(nil)},
	}
	coordinator := NewLifecycleCoordinatorWithPorts(sequenceClock(
		time.Date(2026, 6, 12, 13, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 12, 13, 1, 0, 0, time.UTC),
		time.Date(2026, 6, 12, 13, 2, 0, 0, time.UTC),
		time.Date(2026, 6, 12, 13, 3, 0, 0, time.UTC),
	), ports)

	result := coordinator.Run(job)

	if result.State != LifecycleStateCompleted {
		t.Fatalf("State = %s, want %s", result.State, LifecycleStateCompleted)
	}
	if result.Failure != nil {
		t.Fatalf("Failure = %#v, want nil", result.Failure)
	}
	assertStringSlice(t, calls, []string{"precheck", "planner", "generation", "result"})
}

func TestCoordinatorMapsDownstreamStageFailuresToFailedTerminalState(t *testing.T) {
	tests := []struct {
		name             string
		failure          LifecycleError
		ports            func(*testing.T, *[]string, LifecycleError) DownstreamPorts
		wantCalls        []string
		wantFailureStage LifecycleStage
		wantFieldPath    string
	}{
		{
			name:             "planner failure stops before generation and result",
			failure:          NewLifecycleError(LifecycleErrorCodeDownstreamFailure, LifecycleStagePlanner, "planner.dependencyPlan", "password=planner-secret host=db.internal:5432 SELECT * FROM users"),
			wantCalls:        []string{"precheck", "planner"},
			wantFailureStage: LifecycleStagePlanner,
			wantFieldPath:    "planner.dependencyPlan",
			ports: func(t *testing.T, calls *[]string, failure LifecycleError) DownstreamPorts {
				return DownstreamPorts{
					Precheck:   fakePrecheckPort{t: t, calls: calls, name: "precheck"},
					Planner:    fakePlannerPort{t: t, calls: calls, name: "planner", result: NewDownstreamStageFailure(failure)},
					Generation: fakeGenerationPort{t: t, calls: calls, name: "generation", result: NewDownstreamStageSuccess("must-not-run")},
					Result:     fakeResultPort{t: t, calls: calls, name: "result", result: NewDownstreamStageSuccess(nil)},
				}
			},
		},
		{
			name:             "generation failure stops before result",
			failure:          NewLifecycleError(LifecycleErrorCodeDownstreamFailure, LifecycleStageGeneration, "generation.batch", "generated data: [{email:'a@example.test'}] token=raw-token"),
			wantCalls:        []string{"precheck", "planner", "generation"},
			wantFailureStage: LifecycleStageGeneration,
			wantFieldPath:    "generation.batch",
			ports: func(t *testing.T, calls *[]string, failure LifecycleError) DownstreamPorts {
				return DownstreamPorts{
					Precheck:   fakePrecheckPort{t: t, calls: calls, name: "precheck"},
					Planner:    fakePlannerPort{t: t, calls: calls, name: "planner", result: NewDownstreamStageSuccess("opaque-plan")},
					Generation: fakeGenerationPort{t: t, calls: calls, name: "generation", wantPlanArtifact: "opaque-plan", result: NewDownstreamStageFailure(failure)},
					Result:     fakeResultPort{t: t, calls: calls, name: "result", result: NewDownstreamStageSuccess(nil)},
				}
			},
		},
		{
			name:             "result failure becomes failed terminal state",
			failure:          NewLifecycleError(LifecycleErrorCodeDownstreamFailure, LifecycleStageResult, "result.summary", "connection string password=summary-secret user SQL DELETE FROM users"),
			wantCalls:        []string{"precheck", "planner", "generation", "result"},
			wantFailureStage: LifecycleStageResult,
			wantFieldPath:    "result.summary",
			ports: func(t *testing.T, calls *[]string, failure LifecycleError) DownstreamPorts {
				return DownstreamPorts{
					Precheck:   fakePrecheckPort{t: t, calls: calls, name: "precheck"},
					Planner:    fakePlannerPort{t: t, calls: calls, name: "planner", result: NewDownstreamStageSuccess("opaque-plan")},
					Generation: fakeGenerationPort{t: t, calls: calls, name: "generation", wantPlanArtifact: "opaque-plan", result: NewDownstreamStageSuccess("opaque-generation")},
					Result:     fakeResultPort{t: t, calls: calls, name: "result", wantPlanArtifact: "opaque-plan", wantGenerationArtifact: "opaque-generation", result: NewDownstreamStageFailure(failure)},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			coordinator := NewLifecycleCoordinatorWithPorts(sequenceClock(
				time.Date(2026, 6, 12, 16, 0, 0, 0, time.UTC),
				time.Date(2026, 6, 12, 16, 1, 0, 0, time.UTC),
				time.Date(2026, 6, 12, 16, 2, 0, 0, time.UTC),
				time.Date(2026, 6, 12, 16, 3, 0, 0, time.UTC),
			), tt.ports(t, &calls, tt.failure))

			result := coordinator.Run(validGenerationJobSnapshot())

			if result.State != LifecycleStateFailed {
				t.Fatalf("State = %s, want %s", result.State, LifecycleStateFailed)
			}
			if result.StartedAt == nil {
				t.Fatal("StartedAt = nil, want runtime start time before downstream failure")
			}
			if result.CompletedAt != nil {
				t.Fatalf("CompletedAt = %v, want nil", result.CompletedAt)
			}
			assertDownstreamFailureSummary(t, result.Failure, tt.wantFailureStage, tt.wantFieldPath)
			if !result.Snapshot.EndedAt.Equal(result.Transitions[len(result.Transitions)-1].OccurredAt) {
				t.Fatalf("Snapshot.EndedAt = %v, want failed transition time %s", result.Snapshot.EndedAt, result.Transitions[len(result.Transitions)-1].OccurredAt)
			}
			assertDownstreamFailureSummary(t, result.Snapshot.Failure, tt.wantFailureStage, tt.wantFieldPath)
			if result.Snapshot.CancellationRequested {
				t.Fatal("failed downstream lifecycle should not mark cancellation intent")
			}
			assertStringSlice(t, calls, tt.wantCalls)
			assertTransitionPath(t, result.Transitions, []LifecycleState{
				LifecycleStateInitialized,
				LifecycleStatePrechecking,
				LifecycleStateReady,
				LifecycleStateRunning,
				LifecycleStateFailed,
			})
		})
	}
}

func TestCoordinatorSanitizesMalformedDownstreamFailureSummary(t *testing.T) {
	var calls []string
	failure := LifecycleError{
		Code:        "RAW_DRIVER_CODE password=hunter2",
		Stage:       LifecycleStageGeneration,
		FieldPath:   "generation.raw password=hunter2 host=db.internal SELECT * FROM users generated data",
		SafeMessage: "password=hunter2 host=db.internal:5432 SELECT * FROM users generated data: [{email:'a@example.test'}]",
	}
	ports := DownstreamPorts{
		Precheck:   fakePrecheckPort{calls: &calls, name: "precheck"},
		Planner:    fakePlannerPort{calls: &calls, name: "planner", result: NewDownstreamStageSuccess("opaque-plan")},
		Generation: fakeGenerationPort{calls: &calls, name: "generation", result: DownstreamStageResult{Status: DownstreamStageFailed, Failure: &failure}},
		Result:     fakeResultPort{calls: &calls, name: "result", result: NewDownstreamStageSuccess(nil)},
	}
	coordinator := NewLifecycleCoordinatorWithPorts(sequenceClock(
		time.Date(2026, 6, 12, 17, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 12, 17, 1, 0, 0, time.UTC),
		time.Date(2026, 6, 12, 17, 2, 0, 0, time.UTC),
		time.Date(2026, 6, 12, 17, 3, 0, 0, time.UTC),
	), ports)

	result := coordinator.Run(validGenerationJobSnapshot())

	if result.State != LifecycleStateFailed {
		t.Fatalf("State = %s, want %s", result.State, LifecycleStateFailed)
	}
	if result.CompletedAt != nil {
		t.Fatalf("CompletedAt = %v, want nil", result.CompletedAt)
	}
	assertDownstreamFailureSummary(t, result.Failure, LifecycleStageGeneration, "generation")
	assertNoSensitiveLifecycleErrorContent(t, *result.Failure)
	if result.Failure.Code != LifecycleErrorCodeDownstreamFailure {
		t.Fatalf("Failure.Code = %s, want %s", result.Failure.Code, LifecycleErrorCodeDownstreamFailure)
	}
	if strings.Contains(result.Failure.Code.String(), "password") {
		t.Fatalf("Failure.Code leaked raw downstream content: %q", result.Failure.Code)
	}
	assertStringSlice(t, calls, []string{"precheck", "planner", "generation"})
}

func assertDownstreamFailureSummary(t *testing.T, failure *LifecycleError, stage LifecycleStage, fieldPath string) {
	t.Helper()
	if failure == nil {
		t.Fatal("Failure = nil, want safe downstream failure summary")
	}
	if failure.Code != LifecycleErrorCodeDownstreamFailure {
		t.Fatalf("Failure.Code = %s, want %s", failure.Code, LifecycleErrorCodeDownstreamFailure)
	}
	if failure.Stage != stage {
		t.Fatalf("Failure.Stage = %s, want %s", failure.Stage, stage)
	}
	if failure.FieldPath != fieldPath {
		t.Fatalf("Failure.FieldPath = %q, want %q", failure.FieldPath, fieldPath)
	}
	if failure.SafeMessage != "downstream lifecycle stage failed" {
		t.Fatalf("Failure.SafeMessage = %q, want generic downstream safe message", failure.SafeMessage)
	}
	assertNoSensitiveFragments(t, failure.SafeMessage)
}

func TestDownstreamStageFailureCarriesOnlySafeError(t *testing.T) {
	failure := NewLifecycleError(LifecycleErrorCodeDownstreamFailure, LifecycleStageGeneration, "generation", "password=secret select * from users")
	result := NewDownstreamStageFailure(failure)

	if result.Status != DownstreamStageFailed {
		t.Fatalf("Status = %s, want %s", result.Status, DownstreamStageFailed)
	}
	if result.Failure == nil {
		t.Fatal("Failure = nil, want safe lifecycle error")
	}
	if result.Failure.Code != LifecycleErrorCodeDownstreamFailure {
		t.Fatalf("Failure.Code = %s, want %s", result.Failure.Code, LifecycleErrorCodeDownstreamFailure)
	}
	if result.Failure.Stage != LifecycleStageGeneration {
		t.Fatalf("Failure.Stage = %s, want %s", result.Failure.Stage, LifecycleStageGeneration)
	}
	if result.Failure.SafeMessage != "downstream lifecycle stage failed" {
		t.Fatalf("Failure.SafeMessage = %q, want generic safe message", result.Failure.SafeMessage)
	}
}

func TestDownstreamPrecheckResultIsAggregatedBeforeStart(t *testing.T) {
	failure := NewLifecycleError(LifecycleErrorCodeInvalidReference, LifecycleStagePrecheck, "planner.schema", "planner prerequisite is not available")
	ports := DownstreamPorts{
		Precheck: fakePrecheckPort{result: PrecheckResult{BlockingErrors: []PrecheckIssue{failure}}},
		Planner:  fakePlannerPort{result: NewDownstreamStageSuccess(nil)},
	}
	coordinator := NewLifecycleCoordinatorWithPorts(sequenceClock(time.Date(2026, 6, 12, 15, 0, 0, 0, time.UTC)), ports)

	result := coordinator.Run(validGenerationJobSnapshot())

	if result.Precheck.Passed {
		t.Fatal("Precheck.Passed = true, want false")
	}
	if result.State != LifecycleStateFailed {
		t.Fatalf("State = %s, want %s", result.State, LifecycleStateFailed)
	}
	if result.StartedAt != nil {
		t.Fatalf("StartedAt = %v, want nil", result.StartedAt)
	}
	assertBlockingIssue(t, result.Precheck, "planner.schema", LifecycleErrorCodeInvalidReference)
}

func TestDownstreamPortsStayMinimalAndOpaque(t *testing.T) {
	resultType := reflect.TypeFor[DownstreamStageResult]()
	assertStructFields(t, resultType, []string{"Status", "Failure", "Artifact"})
	if resultType.Field(2).Type.Kind() != reflect.Interface {
		t.Fatalf("Artifact kind = %s, want opaque interface", resultType.Field(2).Type.Kind())
	}

	contextType := reflect.TypeFor[DownstreamContext]()
	assertStructFields(t, contextType, []string{"Input", "Control", "PlanArtifact", "GenerationArtifact"})
	for _, disallowed := range []string{"DependencyGraph", "RowCountPlan", "GeneratorRegistry", "BatchLoop", "ResultAggregator", "WriterAdapter", "Database"} {
		if _, ok := resultType.FieldByName(disallowed); ok {
			t.Fatalf("DownstreamStageResult must not expose future algorithm field %s", disallowed)
		}
		if _, ok := contextType.FieldByName(disallowed); ok {
			t.Fatalf("DownstreamContext must not expose future algorithm field %s", disallowed)
		}
	}
}

type fakePrecheckPort struct {
	t      *testing.T
	calls  *[]string
	name   string
	result PrecheckResult
}

func (p fakePrecheckPort) Precheck(context DownstreamContext) PrecheckResult {
	if p.calls != nil {
		*p.calls = append(*p.calls, p.name)
	}
	if p.t != nil {
		p.assertMinimalContext(context)
	}
	return p.result
}

func (p fakePrecheckPort) assertMinimalContext(context DownstreamContext) {
	p.t.Helper()
	if context.Input == nil {
		p.t.Fatal("downstream precheck should receive execution input")
	}
	if context.Control.CancellationRequested() {
		p.t.Fatal("downstream precheck should receive non-cancelled control token")
	}
}

type fakePlannerPort struct {
	t      *testing.T
	calls  *[]string
	name   string
	result DownstreamStageResult
}

func (p fakePlannerPort) Plan(context DownstreamContext) DownstreamStageResult {
	if p.calls != nil {
		*p.calls = append(*p.calls, p.name)
	}
	if p.t != nil {
		p.t.Helper()
		if context.Input == nil {
			p.t.Fatal("planner should receive execution input")
		}
		if context.PlanArtifact != nil || context.GenerationArtifact != nil {
			p.t.Fatalf("planner context should not precompute artifacts, got plan=%#v generation=%#v", context.PlanArtifact, context.GenerationArtifact)
		}
	}
	return p.result
}

type fakeGenerationPort struct {
	t                      *testing.T
	calls                  *[]string
	name                   string
	wantPlanArtifact       any
	wantGenerationArtifact any
	result                 DownstreamStageResult
}

func (p fakeGenerationPort) Generate(context DownstreamContext) DownstreamStageResult {
	if p.calls != nil {
		*p.calls = append(*p.calls, p.name)
	}
	if p.t != nil {
		p.t.Helper()
		if context.Input == nil {
			p.t.Fatal("generation should receive execution input")
		}
		if context.PlanArtifact != p.wantPlanArtifact {
			p.t.Fatalf("PlanArtifact = %#v, want %#v", context.PlanArtifact, p.wantPlanArtifact)
		}
		if context.GenerationArtifact != p.wantGenerationArtifact {
			p.t.Fatalf("GenerationArtifact = %#v, want %#v", context.GenerationArtifact, p.wantGenerationArtifact)
		}
	}
	return p.result
}

type fakeResultPort struct {
	t                      *testing.T
	calls                  *[]string
	name                   string
	wantPlanArtifact       any
	wantGenerationArtifact any
	result                 DownstreamStageResult
}

func (p fakeResultPort) Summarize(context DownstreamContext) DownstreamStageResult {
	if p.calls != nil {
		*p.calls = append(*p.calls, p.name)
	}
	if p.t != nil {
		p.t.Helper()
		if context.Input == nil {
			p.t.Fatal("result summary should receive execution input")
		}
		if context.PlanArtifact != p.wantPlanArtifact {
			p.t.Fatalf("PlanArtifact = %#v, want %#v", context.PlanArtifact, p.wantPlanArtifact)
		}
		if context.GenerationArtifact != p.wantGenerationArtifact {
			p.t.Fatalf("GenerationArtifact = %#v, want %#v", context.GenerationArtifact, p.wantGenerationArtifact)
		}
	}
	return p.result
}

func assertStringSlice(t *testing.T, got []string, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("calls = %#v, want %#v", got, want)
	}
}

func assertStructFields(t *testing.T, typ reflect.Type, want []string) {
	t.Helper()
	if typ.NumField() != len(want) {
		t.Fatalf("%s field count = %d, want %d", typ.Name(), typ.NumField(), len(want))
	}
	for index, name := range want {
		if typ.Field(index).Name != name {
			t.Fatalf("%s field[%d] = %s, want %s", typ.Name(), index, typ.Field(index).Name, name)
		}
	}
}
