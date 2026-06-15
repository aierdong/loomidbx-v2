package plan

import (
	"testing"

	domainproject "github.com/gerdong/loomidbx/internal/domain/project"
	domainschema "github.com/gerdong/loomidbx/internal/domain/schema"
	"github.com/gerdong/loomidbx/internal/engine/lifecycle"
)

func TestPlanDependenciesBuildsSuccessfulExecutionPlan(t *testing.T) {
	projectSnapshot := domainproject.Project{ID: 101, Name: "demo"}
	tables := []domainproject.ProjectTable{
		{ID: 302, ProjectID: 101, TableID: 202, ExecutionOrder: 2},
		{ID: 301, ProjectID: 101, TableID: 201, ExecutionOrder: 1},
	}
	foreignKeys := []domainschema.ForeignKey{{ID: 401, TableID: 202, ReferencedTableID: 201}}

	result := PlanDependencies(projectSnapshot, tables, foreignKeys, nil, nil)

	if !result.Passed {
		t.Fatalf("expected dependency plan to pass, got %#v", result.BlockingErrors)
	}
	if result.Plan == nil {
		t.Fatal("expected execution plan")
	}
	if result.Plan.ProjectID != 101 {
		t.Fatalf("ProjectID = %d, want 101", result.Plan.ProjectID)
	}
	assertPlannedTableOrder(t, result.Plan.OrderedTables, []PlannedTable{
		{ProjectTableID: 301, TableID: 201, ExecutionOrder: 1},
		{ProjectTableID: 302, TableID: 202, ExecutionOrder: 2},
	})
	if len(result.Plan.Edges) != 1 {
		t.Fatalf("Edges length = %d, want 1", len(result.Plan.Edges))
	}
	if result.Plan.Edges[0].CanonicalKey() != "301->302" {
		t.Fatalf("edge key = %q, want 301->302", result.Plan.Edges[0].CanonicalKey())
	}
}

func TestPlanDependenciesKeepsWarningsWithoutBlockingSuccess(t *testing.T) {
	projectSnapshot := domainproject.Project{ID: 101, Name: "demo"}
	tables := []domainproject.ProjectTable{
		{ID: 301, ProjectID: 101, TableID: 201, ExecutionOrder: 1},
		{ID: 302, ProjectID: 101, TableID: 202, ExecutionOrder: 2},
	}
	foreignKeys := []domainschema.ForeignKey{
		{ID: 401, TableID: 202, ReferencedTableID: 201},
		{ID: 402, TableID: 202, ReferencedTableID: 201},
	}

	result := PlanDependencies(projectSnapshot, tables, foreignKeys, nil, nil)

	if !result.Passed {
		t.Fatalf("expected duplicate-edge warning to pass, got %#v", result.BlockingErrors)
	}
	if result.Plan == nil {
		t.Fatal("expected execution plan")
	}
	if len(result.BlockingErrors) != 0 {
		t.Fatalf("BlockingErrors length = %d, want 0", len(result.BlockingErrors))
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("Warnings length = %d, want 1: %#v", len(result.Warnings), result.Warnings)
	}
	if len(result.Plan.Warnings) != 1 {
		t.Fatalf("plan Warnings length = %d, want 1", len(result.Plan.Warnings))
	}
	if result.Warnings[0].Blocking {
		t.Fatalf("warning was marked blocking: %#v", result.Warnings[0])
	}
	if result.Warnings[0].Code != PlanErrorCodeDuplicateEdge || result.Warnings[0].Stage != PlanStageGraphBuild || result.Warnings[0].FieldPath != "edges[301->302]" {
		t.Fatalf("unexpected warning issue: %#v", result.Warnings[0])
	}
}

func TestPlanDependenciesReturnsSafeFailureWithoutPlan(t *testing.T) {
	projectSnapshot := domainproject.Project{ID: 101, Name: "demo"}
	tables := []domainproject.ProjectTable{
		{ID: 301, ProjectID: 101, TableID: 201, ExecutionOrder: 1},
		{ID: 302, ProjectID: 101, TableID: 202, ExecutionOrder: 2},
	}
	tableRelations := []domainschema.TableRelation{
		{ID: 501, RelationType: domainschema.RelationTypeParentChild, ParentTableID: 201, ChildTableID: 202},
		{ID: 502, RelationType: domainschema.RelationTypeParentChild, ParentTableID: 202, ChildTableID: 201},
	}

	result := PlanDependencies(projectSnapshot, tables, nil, tableRelations, nil)

	if result.Passed {
		t.Fatal("expected cycle to fail dependency plan")
	}
	if result.Plan != nil {
		t.Fatalf("expected no execution plan, got %#v", result.Plan)
	}
	assertPlanIssue(t, result.BlockingErrors, "graph.cycles[301,302]", PlanErrorCodeCycleDetected, PlanStageTopologicalSort)
	assertPlanIssue(t, result.BlockingErrors, "graph.nodes", PlanErrorCodeUnsortableGraph, PlanStageTopologicalSort)
	for _, issue := range result.BlockingErrors {
		assertNoSensitivePlanIssueContent(t, issue)
	}
}

func TestPlanDependenciesFailsSafelyForUnknownRelationInputsWithoutPlan(t *testing.T) {
	parentProjectTableID := int64(301)
	projectSnapshot := domainproject.Project{ID: 101, Name: "demo"}
	tables := []domainproject.ProjectTable{
		{ID: 301, ProjectID: 101, TableID: 201, ExecutionOrder: 1},
		{ID: 302, ProjectID: 101, TableID: 202, ExecutionOrder: 2},
	}
	tableRelations := []domainschema.TableRelation{
		{ID: 501, RelationType: domainschema.RelationType("HAS_MANY"), ParentTableID: 201, ChildTableID: 202},
	}
	projectRelations := []domainproject.ProjectTableRelation{
		{
			ID:                   601,
			ProjectID:            101,
			TableRelationID:      501,
			ParentProjectTableID: &parentProjectTableID,
			ChildProjectTableID:  302,
			RelValueSource:       domainproject.RelationValueSource("FROM_USER_SQL"),
			RelSourceSQL:         "select password from users where host='db.internal'",
		},
	}

	result := PlanDependencies(projectSnapshot, tables, nil, tableRelations, projectRelations)

	if result.Passed {
		t.Fatal("expected unknown relation inputs to fail dependency plan")
	}
	if result.Plan != nil {
		t.Fatalf("expected no execution plan, got %#v", result.Plan)
	}
	assertPlanIssue(t, result.BlockingErrors, "tableRelations[0].relationType", PlanErrorCodeUnknownRelationType, PlanStageRelationMapping)
	assertPlanIssue(t, result.BlockingErrors, "projectRelations[0].relValueSource", PlanErrorCodeUnknownValueSource, PlanStageRelationMapping)
	for _, issue := range result.BlockingErrors {
		assertNoSensitivePlanIssueContent(t, issue)
	}
}

func TestPlanDependenciesDoesNotModifyInputRelationConfiguration(t *testing.T) {
	parentProjectTableID := int64(301)
	projectSnapshot := domainproject.Project{ID: 101, Name: "demo"}
	tables := []domainproject.ProjectTable{
		{ID: 301, ProjectID: 101, TableID: 201, ExecutionOrder: 1},
		{ID: 302, ProjectID: 101, TableID: 202, ExecutionOrder: 2},
	}
	foreignKeys := []domainschema.ForeignKey{{ID: 401, TableID: 202, ReferencedTableID: 201, ColumnIDs: []int64{11}, ReferencedColumnIDs: []int64{12}}}
	tableRelations := []domainschema.TableRelation{{ID: 501, RelationType: domainschema.RelationTypeParentChild, ParentTableID: 201, ChildTableID: 202, ParentColumnIDs: []int64{11}, ChildColumnIDs: []int64{12}}}
	projectRelations := []domainproject.ProjectTableRelation{{
		ID:                   601,
		ProjectID:            101,
		TableRelationID:      501,
		ParentProjectTableID: &parentProjectTableID,
		ChildProjectTableID:  302,
		RelValueSource:       domainproject.RelationValueSourceFromExecution,
		RelSourceSQL:         "select password from users where host='db.internal'",
	}}
	originalForeignKeys := append([]domainschema.ForeignKey(nil), foreignKeys...)
	originalTableRelations := append([]domainschema.TableRelation(nil), tableRelations...)
	originalProjectRelations := append([]domainproject.ProjectTableRelation(nil), projectRelations...)

	result := PlanDependencies(projectSnapshot, tables, foreignKeys, tableRelations, projectRelations)

	if !result.Passed {
		t.Fatalf("expected dependency plan to pass, got %#v", result.BlockingErrors)
	}
	assertForeignKeySnapshotsEqual(t, foreignKeys, originalForeignKeys)
	assertTableRelationSnapshotsEqual(t, tableRelations, originalTableRelations)
	assertProjectRelationSnapshotsEqual(t, projectRelations, originalProjectRelations)
}

func TestDependencyPlannerPrecheckMapsPlanIssuesToLifecyclePrecheck(t *testing.T) {
	planner := NewDependencyPlanner(
		domainproject.Project{ID: 101, Name: "demo"},
		[]domainproject.ProjectTable{{ID: 301, ProjectID: 101, TableID: 201, ExecutionOrder: 1}},
		[]domainschema.ForeignKey{{ID: 401, TableID: 202, ReferencedTableID: 201}},
		nil,
		nil,
	)

	precheck := planner.Precheck(lifecycle.DownstreamContext{})

	if precheck.Passed {
		t.Fatal("expected missing endpoint to fail lifecycle precheck")
	}
	if len(precheck.BlockingErrors) != 1 {
		t.Fatalf("BlockingErrors length = %d, want 1: %#v", len(precheck.BlockingErrors), precheck.BlockingErrors)
	}
	issue := precheck.BlockingErrors[0]
	if issue.Code != lifecycle.LifecycleErrorCodeInvalidReference {
		t.Fatalf("Code = %s, want %s", issue.Code, lifecycle.LifecycleErrorCodeInvalidReference)
	}
	if issue.Stage != lifecycle.LifecycleStagePrecheck {
		t.Fatalf("Stage = %s, want %s", issue.Stage, lifecycle.LifecycleStagePrecheck)
	}
	if issue.FieldPath != "foreignKeys[0].tableId" {
		t.Fatalf("FieldPath = %q, want foreignKeys[0].tableId", issue.FieldPath)
	}
	assertNoSensitivePlanFragments(t, issue.SafeMessage)
}

func TestDependencyPlannerLifecyclePlanReturnsExecutionPlanArtifact(t *testing.T) {
	planner := NewDependencyPlanner(
		domainproject.Project{ID: 101, Name: "demo"},
		[]domainproject.ProjectTable{{ID: 301, ProjectID: 101, TableID: 201, ExecutionOrder: 1}},
		nil,
		nil,
		nil,
	)

	result := planner.Plan(lifecycle.DownstreamContext{Input: &lifecycle.ExecutionInput{ProjectID: 101}})

	if result.Status != lifecycle.DownstreamStageSucceeded {
		t.Fatalf("Status = %s, want %s: %#v", result.Status, lifecycle.DownstreamStageSucceeded, result.Failure)
	}
	artifact, ok := result.Artifact.(*ExecutionPlan)
	if !ok {
		t.Fatalf("Artifact = %T, want *ExecutionPlan", result.Artifact)
	}
	assertPlannedTableOrder(t, artifact.OrderedTables, []PlannedTable{{ProjectTableID: 301, TableID: 201, ExecutionOrder: 1}})
}

func TestDependencyPlannerLifecyclePlanReturnsSafeFailure(t *testing.T) {
	planner := NewDependencyPlanner(
		domainproject.Project{ID: 101, Name: "demo"},
		[]domainproject.ProjectTable{{ID: 301, ProjectID: 101, TableID: 201, ExecutionOrder: 1}},
		[]domainschema.ForeignKey{{ID: 401, TableID: 202, ReferencedTableID: 201}},
		nil,
		nil,
	)

	result := planner.Plan(lifecycle.DownstreamContext{Input: &lifecycle.ExecutionInput{ProjectID: 101}})

	if result.Status != lifecycle.DownstreamStageFailed {
		t.Fatalf("Status = %s, want %s", result.Status, lifecycle.DownstreamStageFailed)
	}
	if result.Artifact != nil {
		t.Fatalf("Artifact = %#v, want nil", result.Artifact)
	}
	if result.Failure == nil {
		t.Fatal("expected safe lifecycle failure")
	}
	if result.Failure.Code != lifecycle.LifecycleErrorCodeInvalidReference {
		t.Fatalf("Code = %s, want %s", result.Failure.Code, lifecycle.LifecycleErrorCodeInvalidReference)
	}
	if result.Failure.Stage != lifecycle.LifecycleStagePlanner {
		t.Fatalf("Stage = %s, want %s", result.Failure.Stage, lifecycle.LifecycleStagePlanner)
	}
	if result.Failure.FieldPath != "foreignKeys[0].tableId" {
		t.Fatalf("FieldPath = %q, want foreignKeys[0].tableId", result.Failure.FieldPath)
	}
	assertNoSensitivePlanFragments(t, result.Failure.SafeMessage)
}

func assertForeignKeySnapshotsEqual(t *testing.T, got []domainschema.ForeignKey, want []domainschema.ForeignKey) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("ForeignKey length = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index].ID != want[index].ID || got[index].TableID != want[index].TableID || got[index].ReferencedTableID != want[index].ReferencedTableID || got[index].FKName != want[index].FKName {
			t.Fatalf("ForeignKey[%d] = %#v, want %#v", index, got[index], want[index])
		}
		assertInt64SliceEqual(t, got[index].ColumnIDs, want[index].ColumnIDs)
		assertInt64SliceEqual(t, got[index].ReferencedColumnIDs, want[index].ReferencedColumnIDs)
	}
}

func assertTableRelationSnapshotsEqual(t *testing.T, got []domainschema.TableRelation, want []domainschema.TableRelation) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("TableRelation length = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index].ID != want[index].ID || got[index].RelationType != want[index].RelationType || got[index].ParentTableID != want[index].ParentTableID || got[index].ChildTableID != want[index].ChildTableID || got[index].MultiplierMin != want[index].MultiplierMin || got[index].MultiplierMax != want[index].MultiplierMax || got[index].IsLogical != want[index].IsLogical {
			t.Fatalf("TableRelation[%d] = %#v, want %#v", index, got[index], want[index])
		}
		assertInt64SliceEqual(t, got[index].ParentColumnIDs, want[index].ParentColumnIDs)
		assertInt64SliceEqual(t, got[index].ChildColumnIDs, want[index].ChildColumnIDs)
	}
}

func assertProjectRelationSnapshotsEqual(t *testing.T, got []domainproject.ProjectTableRelation, want []domainproject.ProjectTableRelation) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("ProjectTableRelation length = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index].ID != want[index].ID || got[index].ProjectID != want[index].ProjectID || got[index].TableRelationID != want[index].TableRelationID || got[index].ChildProjectTableID != want[index].ChildProjectTableID || got[index].MultiplierMin != want[index].MultiplierMin || got[index].MultiplierMax != want[index].MultiplierMax || got[index].RelValueSource != want[index].RelValueSource || got[index].RelSourceSQL != want[index].RelSourceSQL {
			t.Fatalf("ProjectTableRelation[%d] = %#v, want %#v", index, got[index], want[index])
		}
		if (got[index].ParentProjectTableID == nil) != (want[index].ParentProjectTableID == nil) {
			t.Fatalf("ProjectTableRelation[%d].ParentProjectTableID = %#v, want %#v", index, got[index].ParentProjectTableID, want[index].ParentProjectTableID)
		}
		if got[index].ParentProjectTableID != nil && *got[index].ParentProjectTableID != *want[index].ParentProjectTableID {
			t.Fatalf("ProjectTableRelation[%d].ParentProjectTableID = %d, want %d", index, *got[index].ParentProjectTableID, *want[index].ParentProjectTableID)
		}
	}
}

func assertInt64SliceEqual(t *testing.T, got []int64, want []int64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("int64 slice length = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("int64 slice[%d] = %d, want %d", index, got[index], want[index])
		}
	}
}
