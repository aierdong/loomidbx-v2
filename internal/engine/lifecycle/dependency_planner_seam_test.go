package lifecycle_test

import (
	"testing"
	"time"

	domainexecution "github.com/gerdong/loomidbx/internal/domain/execution"
	domainproject "github.com/gerdong/loomidbx/internal/domain/project"
	domainschema "github.com/gerdong/loomidbx/internal/domain/schema"
	"github.com/gerdong/loomidbx/internal/engine/lifecycle"
	"github.com/gerdong/loomidbx/internal/engine/plan"
)

func TestDependencyPlannerSeamWarningsDoNotBlockLifecycleStartup(t *testing.T) {
	planner := plan.NewDependencyPlanner(
		domainproject.Project{ID: 101, Name: "demo"},
		[]domainproject.ProjectTable{
			{ID: 301, ProjectID: 101, TableID: 201, ExecutionOrder: 1},
			{ID: 302, ProjectID: 101, TableID: 202, ExecutionOrder: 2},
		},
		[]domainschema.ForeignKey{
			{ID: 401, TableID: 202, ReferencedTableID: 201},
			{ID: 402, TableID: 202, ReferencedTableID: 201},
		},
		nil,
		nil,
	)
	generationSawPlan := false
	ports := lifecycle.NoopDownstreamPorts()
	ports.Precheck = planner
	ports.Planner = planner
	ports.Generation = dependencyPlanGenerationProbe{t: t, sawPlan: &generationSawPlan}
	coordinator := lifecycle.NewLifecycleCoordinatorWithPorts(sequenceClock(
		time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 15, 10, 1, 0, 0, time.UTC),
		time.Date(2026, 6, 15, 10, 2, 0, 0, time.UTC),
		time.Date(2026, 6, 15, 10, 3, 0, 0, time.UTC),
	), ports)

	result := coordinator.Run(validDependencyPlannerJob())

	if !result.Precheck.Passed {
		t.Fatalf("Precheck.Passed = false, blocking=%#v", result.Precheck.BlockingErrors)
	}
	if len(result.Precheck.Warnings) != 1 {
		t.Fatalf("Warnings length = %d, want 1: %#v", len(result.Precheck.Warnings), result.Precheck.Warnings)
	}
	if result.Precheck.Warnings[0].Code != lifecycle.LifecycleErrorCodeDownstreamFailure || result.Precheck.Warnings[0].FieldPath != "edges[301->302]" {
		t.Fatalf("unexpected dependency planner warning: %#v", result.Precheck.Warnings[0])
	}
	if result.State != lifecycle.LifecycleStateCompleted {
		t.Fatalf("State = %s, want %s", result.State, lifecycle.LifecycleStateCompleted)
	}
	if result.StartedAt == nil {
		t.Fatal("StartedAt = nil, want lifecycle to enter RUNNING")
	}
	if !generationSawPlan {
		t.Fatal("generation seam did not receive dependency execution plan artifact")
	}
	assertLifecycleTransitionPath(t, result.Transitions, []lifecycle.LifecycleState{
		lifecycle.LifecycleStateInitialized,
		lifecycle.LifecycleStatePrechecking,
		lifecycle.LifecycleStateReady,
		lifecycle.LifecycleStateRunning,
		lifecycle.LifecycleStateCompleted,
	})
}

func TestDependencyPlannerSeamBlockingErrorsStopLifecycleBeforeRunning(t *testing.T) {
	planner := plan.NewDependencyPlanner(
		domainproject.Project{ID: 101, Name: "demo"},
		[]domainproject.ProjectTable{{ID: 301, ProjectID: 101, TableID: 201, ExecutionOrder: 1}},
		[]domainschema.ForeignKey{{ID: 401, TableID: 202, ReferencedTableID: 201}},
		nil,
		nil,
	)
	ports := lifecycle.NoopDownstreamPorts()
	ports.Precheck = planner
	ports.Planner = panicPlannerPort{t: t}
	coordinator := lifecycle.NewLifecycleCoordinatorWithPorts(sequenceClock(
		time.Date(2026, 6, 15, 11, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 15, 11, 1, 0, 0, time.UTC),
	), ports)

	result := coordinator.Run(validDependencyPlannerJob())

	if result.Precheck.Passed {
		t.Fatal("Precheck.Passed = true, want dependency planner blocking error")
	}
	if len(result.Precheck.BlockingErrors) != 1 {
		t.Fatalf("BlockingErrors length = %d, want 1: %#v", len(result.Precheck.BlockingErrors), result.Precheck.BlockingErrors)
	}
	if result.Precheck.BlockingErrors[0].Code != lifecycle.LifecycleErrorCodeInvalidReference || result.Precheck.BlockingErrors[0].FieldPath != "foreignKeys[0].tableId" {
		t.Fatalf("unexpected dependency planner blocking error: %#v", result.Precheck.BlockingErrors[0])
	}
	if result.State != lifecycle.LifecycleStateFailed {
		t.Fatalf("State = %s, want %s", result.State, lifecycle.LifecycleStateFailed)
	}
	if result.StartedAt != nil {
		t.Fatalf("StartedAt = %v, want nil", result.StartedAt)
	}
	assertLifecycleTransitionPath(t, result.Transitions, []lifecycle.LifecycleState{
		lifecycle.LifecycleStateInitialized,
		lifecycle.LifecycleStatePrechecking,
		lifecycle.LifecycleStateFailed,
	})
}

type dependencyPlanGenerationProbe struct {
	t       *testing.T
	sawPlan *bool
}

func (p dependencyPlanGenerationProbe) Generate(context lifecycle.DownstreamContext) lifecycle.DownstreamStageResult {
	p.t.Helper()
	artifact, ok := context.PlanArtifact.(*plan.ExecutionPlan)
	if !ok {
		p.t.Fatalf("PlanArtifact = %T, want *plan.ExecutionPlan", context.PlanArtifact)
	}
	if artifact.ProjectID != 101 {
		p.t.Fatalf("ProjectID = %d, want 101", artifact.ProjectID)
	}
	if len(artifact.OrderedTables) != 2 || artifact.OrderedTables[0].ProjectTableID != 301 || artifact.OrderedTables[1].ProjectTableID != 302 {
		p.t.Fatalf("OrderedTables = %#v, want dependency order 301 then 302", artifact.OrderedTables)
	}
	if len(artifact.Warnings) != 1 {
		p.t.Fatalf("plan Warnings length = %d, want 1", len(artifact.Warnings))
	}
	*p.sawPlan = true
	return lifecycle.NewDownstreamStageSuccess("generated")
}

type panicPlannerPort struct {
	t *testing.T
}

func (p panicPlannerPort) Plan(context lifecycle.DownstreamContext) lifecycle.DownstreamStageResult {
	p.t.Fatal("planner seam must not run after dependency precheck blocking errors")
	return lifecycle.NewDownstreamStageSuccess(nil)
}

func validDependencyPlannerJob() domainexecution.GenerationJob {
	tableID := int64(201)
	return domainexecution.GenerationJob{
		Task: domainexecution.ExecutionTask{
			ID:        701,
			ProjectID: 101,
			TaskName:  "dependency planner seam",
		},
		TableResults: []domainexecution.ExecutionTableResult{{
			ID:                 801,
			ExecutionTaskID:    701,
			TableID:            &tableID,
			SchemaNameSnapshot: "public",
			TableNameSnapshot:  "users",
			ExecutionOrder:     1,
		}},
	}
}

func sequenceClock(times ...time.Time) func() time.Time {
	index := 0
	return func() time.Time {
		if index >= len(times) {
			return times[len(times)-1]
		}
		value := times[index]
		index++
		return value
	}
}

func assertLifecycleTransitionPath(t *testing.T, records []lifecycle.TransitionRecord, states []lifecycle.LifecycleState) {
	t.Helper()
	if len(records) != len(states)-1 {
		t.Fatalf("transition count = %d, want %d", len(records), len(states)-1)
	}
	for index, record := range records {
		if record.From != states[index] || record.To != states[index+1] || record.Rejected {
			t.Fatalf("record[%d] = %s -> %s rejected=%t, want %s -> %s accepted", index, record.From, record.To, record.Rejected, states[index], states[index+1])
		}
	}
}
