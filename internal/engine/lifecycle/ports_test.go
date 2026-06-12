package lifecycle

import (
	"reflect"
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
	resultType := reflect.TypeOf(DownstreamStageResult{})
	assertStructFields(t, resultType, []string{"Status", "Failure", "Artifact"})
	if resultType.Field(2).Type.Kind() != reflect.Interface {
		t.Fatalf("Artifact kind = %s, want opaque interface", resultType.Field(2).Type.Kind())
	}

	contextType := reflect.TypeOf(DownstreamContext{})
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
