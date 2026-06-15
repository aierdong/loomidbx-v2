package plan

import "testing"

func TestTopologicalSortOrdersUpstreamBeforeDownstream(t *testing.T) {
	graph := DependencyGraph{
		Nodes: []GraphNode{
			{ProjectTableID: 303, TableID: 203, StableOrder: 3},
			{ProjectTableID: 302, TableID: 202, StableOrder: 2},
			{ProjectTableID: 301, TableID: 201, StableOrder: 1},
		},
		Edges: []DependencyEdge{
			{FromProjectTableID: 301, ToProjectTableID: 302},
			{FromProjectTableID: 302, ToProjectTableID: 303},
		},
	}

	ordered, issues := SortDependencyGraph(graph)

	if len(issues) != 0 {
		t.Fatalf("Issues length = %d, want 0: %#v", len(issues), issues)
	}
	assertPlannedTableOrder(t, ordered, []PlannedTable{
		{ProjectTableID: 301, TableID: 201, ExecutionOrder: 1},
		{ProjectTableID: 302, TableID: 202, ExecutionOrder: 2},
		{ProjectTableID: 303, TableID: 203, ExecutionOrder: 3},
	})
}

func TestTopologicalSortUsesStableKeysForMultipleZeroIndegreeNodes(t *testing.T) {
	graph := DependencyGraph{
		Nodes: []GraphNode{
			{ProjectTableID: 304, TableID: 204, StableOrder: 0},
			{ProjectTableID: 303, TableID: 203, StableOrder: 2},
			{ProjectTableID: 302, TableID: 202, StableOrder: 1},
			{ProjectTableID: 301, TableID: 201, StableOrder: 0},
		},
	}

	ordered, issues := SortDependencyGraph(graph)

	if len(issues) != 0 {
		t.Fatalf("Issues length = %d, want 0: %#v", len(issues), issues)
	}
	assertPlannedTableOrder(t, ordered, []PlannedTable{
		{ProjectTableID: 302, TableID: 202, ExecutionOrder: 1},
		{ProjectTableID: 303, TableID: 203, ExecutionOrder: 2},
		{ProjectTableID: 301, TableID: 201, ExecutionOrder: 3},
		{ProjectTableID: 304, TableID: 204, ExecutionOrder: 4},
	})
}

func TestTopologicalSortIncludesIsolatedNodesWithBranchDependencies(t *testing.T) {
	graph := DependencyGraph{
		Nodes: []GraphNode{
			{ProjectTableID: 305, TableID: 205, StableOrder: 5},
			{ProjectTableID: 304, TableID: 204, StableOrder: 4},
			{ProjectTableID: 303, TableID: 203, StableOrder: 3},
			{ProjectTableID: 302, TableID: 202, StableOrder: 2},
			{ProjectTableID: 301, TableID: 201, StableOrder: 1},
		},
		Edges: []DependencyEdge{
			{FromProjectTableID: 301, ToProjectTableID: 303},
			{FromProjectTableID: 302, ToProjectTableID: 303},
			{FromProjectTableID: 303, ToProjectTableID: 305},
		},
	}

	ordered, issues := SortDependencyGraph(graph)

	if len(issues) != 0 {
		t.Fatalf("Issues length = %d, want 0: %#v", len(issues), issues)
	}
	assertPlannedTableOrder(t, ordered, []PlannedTable{
		{ProjectTableID: 301, TableID: 201, ExecutionOrder: 1},
		{ProjectTableID: 302, TableID: 202, ExecutionOrder: 2},
		{ProjectTableID: 303, TableID: 203, ExecutionOrder: 3},
		{ProjectTableID: 304, TableID: 204, ExecutionOrder: 4},
		{ProjectTableID: 305, TableID: 205, ExecutionOrder: 5},
	})
}

func TestTopologicalSortPlacesEveryEdgeUpstreamBeforeDownstream(t *testing.T) {
	graph := DependencyGraph{
		Nodes: []GraphNode{
			{ProjectTableID: 305, TableID: 205, StableOrder: 5},
			{ProjectTableID: 304, TableID: 204, StableOrder: 4},
			{ProjectTableID: 303, TableID: 203, StableOrder: 3},
			{ProjectTableID: 302, TableID: 202, StableOrder: 2},
			{ProjectTableID: 301, TableID: 201, StableOrder: 1},
		},
		Edges: []DependencyEdge{
			{FromProjectTableID: 301, ToProjectTableID: 303},
			{FromProjectTableID: 302, ToProjectTableID: 303},
			{FromProjectTableID: 303, ToProjectTableID: 305},
			{FromProjectTableID: 304, ToProjectTableID: 305},
		},
	}

	ordered, issues := SortDependencyGraph(graph)

	if len(issues) != 0 {
		t.Fatalf("Issues length = %d, want 0: %#v", len(issues), issues)
	}
	ordersByProjectTableID := make(map[int64]int, len(ordered))
	for _, table := range ordered {
		ordersByProjectTableID[table.ProjectTableID] = table.ExecutionOrder
	}
	for _, edge := range graph.Edges {
		fromOrder := ordersByProjectTableID[edge.FromProjectTableID]
		toOrder := ordersByProjectTableID[edge.ToProjectTableID]
		if fromOrder == 0 || toOrder == 0 {
			t.Fatalf("edge %s references a table missing from ordered result %#v", edge.CanonicalKey(), ordered)
		}
		if fromOrder >= toOrder {
			t.Fatalf("edge %s ordered upstream at %d and downstream at %d", edge.CanonicalKey(), fromOrder, toOrder)
		}
	}
}

func TestTopologicalSortIsDeterministicAcrossInputOrder(t *testing.T) {
	first := DependencyGraph{
		Nodes: []GraphNode{
			{ProjectTableID: 303, TableID: 203, StableOrder: 0},
			{ProjectTableID: 301, TableID: 201, StableOrder: 0},
			{ProjectTableID: 302, TableID: 202, StableOrder: 0},
		},
		Edges: []DependencyEdge{{FromProjectTableID: 301, ToProjectTableID: 303}},
	}
	second := DependencyGraph{
		Nodes: []GraphNode{
			{ProjectTableID: 302, TableID: 202, StableOrder: 0},
			{ProjectTableID: 303, TableID: 203, StableOrder: 0},
			{ProjectTableID: 301, TableID: 201, StableOrder: 0},
		},
		Edges: []DependencyEdge{{FromProjectTableID: 301, ToProjectTableID: 303}},
	}

	firstOrdered, firstIssues := SortDependencyGraph(first)
	secondOrdered, secondIssues := SortDependencyGraph(second)

	if len(firstIssues) != 0 || len(secondIssues) != 0 {
		t.Fatalf("unexpected issues: first=%#v second=%#v", firstIssues, secondIssues)
	}
	assertPlannedTableOrder(t, firstOrdered, secondOrdered)
}

func TestTopologicalSortReportsSimpleCycleSafely(t *testing.T) {
	graph := DependencyGraph{
		Nodes: []GraphNode{
			{ProjectTableID: 301, TableID: 201, StableOrder: 1},
		},
		Edges: []DependencyEdge{
			{FromProjectTableID: 301, ToProjectTableID: 301},
		},
	}

	ordered, issues := SortDependencyGraph(graph)

	if ordered != nil {
		t.Fatalf("ordered = %#v, want nil", ordered)
	}
	assertPlanIssue(t, issues, "graph.cycles[301]", PlanErrorCodeCycleDetected, PlanStageTopologicalSort)
	assertNoSensitivePlanIssueContent(t, issues[0])
}

func TestTopologicalSortReportsMultiNodeCycleAndUnsortableGraphSafely(t *testing.T) {
	graph := DependencyGraph{
		Nodes: []GraphNode{
			{ProjectTableID: 301, TableID: 201, StableOrder: 1},
			{ProjectTableID: 302, TableID: 202, StableOrder: 2},
			{ProjectTableID: 303, TableID: 203, StableOrder: 3},
			{ProjectTableID: 304, TableID: 204, StableOrder: 4},
		},
		Edges: []DependencyEdge{
			{FromProjectTableID: 301, ToProjectTableID: 302},
			{FromProjectTableID: 302, ToProjectTableID: 303},
			{FromProjectTableID: 303, ToProjectTableID: 301},
		},
	}

	ordered, issues := SortDependencyGraph(graph)

	if ordered != nil {
		t.Fatalf("ordered = %#v, want nil", ordered)
	}
	assertPlanIssue(t, issues, "graph.cycles[301,302,303]", PlanErrorCodeCycleDetected, PlanStageTopologicalSort)
	assertPlanIssue(t, issues, "graph.nodes", PlanErrorCodeUnsortableGraph, PlanStageTopologicalSort)
	for _, issue := range issues {
		assertNoSensitivePlanIssueContent(t, issue)
	}
}

func TestTopologicalSortDoesNotReportBlockedDownstreamNodeAsCycle(t *testing.T) {
	graph := DependencyGraph{
		Nodes: []GraphNode{
			{ProjectTableID: 301, TableID: 201, StableOrder: 1},
			{ProjectTableID: 302, TableID: 202, StableOrder: 2},
			{ProjectTableID: 303, TableID: 203, StableOrder: 3},
		},
		Edges: []DependencyEdge{
			{FromProjectTableID: 301, ToProjectTableID: 302},
			{FromProjectTableID: 302, ToProjectTableID: 301},
			{FromProjectTableID: 302, ToProjectTableID: 303},
		},
	}

	_, issues := SortDependencyGraph(graph)

	assertPlanIssue(t, issues, "graph.cycles[301,302]", PlanErrorCodeCycleDetected, PlanStageTopologicalSort)
	for _, issue := range issues {
		if issue.FieldPath == "graph.cycles[301,302,303]" {
			t.Fatalf("blocked downstream node was reported as a cycle candidate: %#v", issues)
		}
	}
}

func assertPlannedTableOrder(t *testing.T, got []PlannedTable, want []PlannedTable) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("planned table length = %d, want %d: %#v", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("planned table[%d] = %#v, want %#v", index, got[index], want[index])
		}
	}
}
