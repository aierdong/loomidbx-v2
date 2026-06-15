package plan

import (
	"testing"

	domainproject "github.com/gerdong/loomidbx/internal/domain/project"
	domainschema "github.com/gerdong/loomidbx/internal/domain/schema"
)

func TestDependencyGraphMapsPhysicalForeignKeyToReferencedTableDependency(t *testing.T) {
	input := &PlanInput{
		Tables: []PlanTableInput{
			{ProjectTableID: 301, TableID: 201},
			{ProjectTableID: 302, TableID: 202},
		},
		ForeignKeys: []domainschema.ForeignKey{
			{ID: 401, TableID: 202, ReferencedTableID: 201},
		},
	}

	graph := NewDependencyGraph(input)

	if len(graph.Issues) != 0 {
		t.Fatalf("Issues length = %d, want 0: %#v", len(graph.Issues), graph.Issues)
	}
	if len(graph.Edges) != 1 {
		t.Fatalf("Edges length = %d, want 1", len(graph.Edges))
	}
	edge := graph.Edges[0]
	if edge.FromProjectTableID != 301 || edge.ToProjectTableID != 302 {
		t.Fatalf("edge endpoints = %d -> %d, want 301 -> 302", edge.FromProjectTableID, edge.ToProjectTableID)
	}
	if len(edge.Sources) != 1 {
		t.Fatalf("Sources length = %d, want 1", len(edge.Sources))
	}
	assertEdgeSource(t, edge.Sources[0], EdgeSource{
		Type:     EdgeSourceTypeForeignKey,
		SourceID: 401,
		Summary:  "physical foreign key dependency",
	})
}

func TestDependencyGraphReportsMissingPhysicalForeignKeyEndpoints(t *testing.T) {
	input := &PlanInput{
		Tables: []PlanTableInput{
			{ProjectTableID: 302, TableID: 202},
		},
		ForeignKeys: []domainschema.ForeignKey{
			{ID: 401, TableID: 202, ReferencedTableID: 201},
			{ID: 402, TableID: 203, ReferencedTableID: 202},
		},
	}

	graph := NewDependencyGraph(input)

	if len(graph.Edges) != 0 {
		t.Fatalf("Edges length = %d, want 0", len(graph.Edges))
	}
	assertPlanIssue(t, graph.Issues, "foreignKeys[0].referencedTableId", PlanErrorCodeMissingEndpoint, PlanStageRelationMapping)
	assertPlanIssue(t, graph.Issues, "foreignKeys[1].tableId", PlanErrorCodeMissingEndpoint, PlanStageRelationMapping)
}

func TestDependencyGraphMapsLogicalTableRelationsToParentChildDependency(t *testing.T) {
	input := &PlanInput{
		Tables: []PlanTableInput{
			{ProjectTableID: 301, TableID: 201},
			{ProjectTableID: 302, TableID: 202},
			{ProjectTableID: 303, TableID: 203},
		},
		TableRelations: []domainschema.TableRelation{
			{ID: 501, RelationType: domainschema.RelationTypeParentChild, ParentTableID: 201, ChildTableID: 202},
			{ID: 502, RelationType: domainschema.RelationTypeJoinTable, ParentTableID: 201, ChildTableID: 203},
		},
	}

	graph := NewDependencyGraph(input)

	if len(graph.Issues) != 0 {
		t.Fatalf("Issues length = %d, want 0: %#v", len(graph.Issues), graph.Issues)
	}
	if len(graph.Edges) != 2 {
		t.Fatalf("Edges length = %d, want 2", len(graph.Edges))
	}
	assertDependencyEdge(t, graph.Edges[0], 301, 302, EdgeSource{
		Type:         EdgeSourceTypeTableRelation,
		SourceID:     501,
		RelationType: domainschema.RelationTypeParentChild,
		Summary:      "schema table relation dependency",
	})
	assertDependencyEdge(t, graph.Edges[1], 301, 303, EdgeSource{
		Type:         EdgeSourceTypeTableRelation,
		SourceID:     502,
		RelationType: domainschema.RelationTypeJoinTable,
		Summary:      "schema table relation dependency",
	})
}

func TestDependencyGraphRejectsUnknownLogicalRelationType(t *testing.T) {
	input := &PlanInput{
		Tables: []PlanTableInput{
			{ProjectTableID: 301, TableID: 201},
			{ProjectTableID: 302, TableID: 202},
		},
		TableRelations: []domainschema.TableRelation{
			{ID: 501, RelationType: domainschema.RelationType("HAS_MANY"), ParentTableID: 201, ChildTableID: 202},
		},
	}

	graph := NewDependencyGraph(input)

	if len(graph.Edges) != 0 {
		t.Fatalf("Edges length = %d, want 0", len(graph.Edges))
	}
	assertPlanIssue(t, graph.Issues, "tableRelations[0].relationType", PlanErrorCodeUnknownRelationType, PlanStageRelationMapping)
}

func TestDependencyGraphMapsProjectRelationFromExecutionToCurrentExecutionDependency(t *testing.T) {
	parentProjectTableID := int64(301)
	input := &PlanInput{
		Tables: []PlanTableInput{
			{ProjectTableID: 301, TableID: 201},
			{ProjectTableID: 302, TableID: 202},
		},
		ProjectRelations: []domainproject.ProjectTableRelation{
			{
				ID:                   601,
				TableRelationID:      501,
				ParentProjectTableID: &parentProjectTableID,
				ChildProjectTableID:  302,
				RelValueSource:       domainproject.RelationValueSourceFromExecution,
			},
		},
	}

	graph := NewDependencyGraph(input)

	if len(graph.Issues) != 0 {
		t.Fatalf("Issues length = %d, want 0: %#v", len(graph.Issues), graph.Issues)
	}
	if len(graph.Edges) != 1 {
		t.Fatalf("Edges length = %d, want 1", len(graph.Edges))
	}
	assertDependencyEdge(t, graph.Edges[0], 301, 302, EdgeSource{
		Type:        EdgeSourceTypeProjectRelation,
		SourceID:    601,
		ValueSource: domainproject.RelationValueSourceFromExecution,
		Summary:     "project relation current execution dependency",
	})
}

func TestDependencyGraphKeepsProjectRelationFromDBQueryAsExternalSourceOnly(t *testing.T) {
	input := &PlanInput{
		Tables: []PlanTableInput{
			{ProjectTableID: 302, TableID: 202},
		},
		ProjectRelations: []domainproject.ProjectTableRelation{
			{
				ID:                  601,
				TableRelationID:     501,
				ChildProjectTableID: 302,
				RelValueSource:      domainproject.RelationValueSourceFromDBQuery,
				RelSourceSQL:        "select password from users where host='db.internal'",
			},
		},
	}

	graph := NewDependencyGraph(input)

	if len(graph.Edges) != 0 {
		t.Fatalf("Edges length = %d, want 0", len(graph.Edges))
	}
	if len(graph.ExternalSources) != 1 {
		t.Fatalf("ExternalSources length = %d, want 1: %#v", len(graph.ExternalSources), graph.ExternalSources)
	}
	assertEdgeSource(t, graph.ExternalSources[0], EdgeSource{
		Type:           EdgeSourceTypeProjectRelation,
		SourceID:       601,
		ValueSource:    domainproject.RelationValueSourceFromDBQuery,
		Summary:        "project relation external value source",
		ExternalSource: true,
	})
	assertNoSensitivePlanFragments(t, graph.ExternalSources[0].Summary)
	if len(graph.Issues) != 0 {
		t.Fatalf("Issues length = %d, want 0: %#v", len(graph.Issues), graph.Issues)
	}
}

func TestDependencyGraphMapsMergedProjectRelationWhenParentExists(t *testing.T) {
	parentProjectTableID := int64(301)
	input := &PlanInput{
		Tables: []PlanTableInput{
			{ProjectTableID: 301, TableID: 201},
			{ProjectTableID: 302, TableID: 202},
		},
		ProjectRelations: []domainproject.ProjectTableRelation{
			{
				ID:                   601,
				TableRelationID:      501,
				ParentProjectTableID: &parentProjectTableID,
				ChildProjectTableID:  302,
				RelValueSource:       domainproject.RelationValueSourceMerged,
				RelSourceSQL:         "select password from users where host='db.internal'",
			},
		},
	}

	graph := NewDependencyGraph(input)

	if len(graph.Issues) != 0 {
		t.Fatalf("Issues length = %d, want 0: %#v", len(graph.Issues), graph.Issues)
	}
	if len(graph.Edges) != 1 {
		t.Fatalf("Edges length = %d, want 1", len(graph.Edges))
	}
	assertDependencyEdge(t, graph.Edges[0], 301, 302, EdgeSource{
		Type:           EdgeSourceTypeProjectRelation,
		SourceID:       601,
		ValueSource:    domainproject.RelationValueSourceMerged,
		Summary:        "project relation merged value dependency",
		ExternalSource: true,
	})
}

func TestDependencyGraphReportsMissingProjectRelationParentForCurrentExecution(t *testing.T) {
	missingParentProjectTableID := int64(399)
	input := &PlanInput{
		Tables: []PlanTableInput{
			{ProjectTableID: 302, TableID: 202},
		},
		ProjectRelations: []domainproject.ProjectTableRelation{
			{
				ID:                   601,
				TableRelationID:      501,
				ParentProjectTableID: &missingParentProjectTableID,
				ChildProjectTableID:  302,
				RelValueSource:       domainproject.RelationValueSourceFromExecution,
			},
			{
				ID:                  602,
				TableRelationID:     502,
				ChildProjectTableID: 302,
				RelValueSource:      domainproject.RelationValueSourceMerged,
			},
		},
	}

	graph := NewDependencyGraph(input)

	if len(graph.Edges) != 0 {
		t.Fatalf("Edges length = %d, want 0", len(graph.Edges))
	}
	assertPlanIssue(t, graph.Issues, "projectRelations[0].parentProjectTableId", PlanErrorCodeMissingEndpoint, PlanStageRelationMapping)
	assertPlanIssue(t, graph.Issues, "projectRelations[1].parentProjectTableId", PlanErrorCodeMissingEndpoint, PlanStageRelationMapping)
}

func TestDependencyGraphReportsMissingProjectRelationChildAndUnknownValueSource(t *testing.T) {
	parentProjectTableID := int64(301)
	input := &PlanInput{
		Tables: []PlanTableInput{
			{ProjectTableID: 301, TableID: 201},
		},
		ProjectRelations: []domainproject.ProjectTableRelation{
			{
				ID:                   601,
				TableRelationID:      501,
				ParentProjectTableID: &parentProjectTableID,
				ChildProjectTableID:  399,
				RelValueSource:       domainproject.RelationValueSourceFromExecution,
			},
			{
				ID:                  602,
				TableRelationID:     502,
				ChildProjectTableID: 302,
				RelValueSource:      domainproject.RelationValueSource("FROM_USER_SQL"),
				RelSourceSQL:        "select password from users where host='db.internal'",
			},
		},
	}

	graph := NewDependencyGraph(input)

	if len(graph.Edges) != 0 {
		t.Fatalf("Edges length = %d, want 0", len(graph.Edges))
	}
	assertPlanIssue(t, graph.Issues, "projectRelations[0].childProjectTableId", PlanErrorCodeMissingEndpoint, PlanStageRelationMapping)
	assertPlanIssue(t, graph.Issues, "projectRelations[1].relValueSource", PlanErrorCodeUnknownValueSource, PlanStageRelationMapping)
	for _, issue := range graph.Issues {
		assertNoSensitivePlanIssueContent(t, issue)
	}
}

func TestNewDependencyGraphMapsPlanTablesToNodes(t *testing.T) {
	input := &PlanInput{Tables: []PlanTableInput{
		{ProjectTableID: 301, TableID: 201, ExistingExecutionOrder: 2},
		{ProjectTableID: 302, TableID: 202, ExistingExecutionOrder: 1},
	}}

	graph := NewDependencyGraph(input)

	if len(graph.Nodes) != 2 {
		t.Fatalf("Nodes length = %d, want 2", len(graph.Nodes))
	}
	assertGraphNode(t, graph.Nodes[0], GraphNode{ProjectTableID: 301, TableID: 201, StableOrder: 2})
	assertGraphNode(t, graph.Nodes[1], GraphNode{ProjectTableID: 302, TableID: 202, StableOrder: 1})
	if len(graph.Edges) != 0 {
		t.Fatalf("Edges length = %d, want 0", len(graph.Edges))
	}
	if len(graph.Issues) != 0 {
		t.Fatalf("Issues length = %d, want 0", len(graph.Issues))
	}
}

func TestDependencyGraphDeduplicatesEdgesByCanonicalKeyAndMergesSources(t *testing.T) {
	graph := DependencyGraph{}
	foreignKeySource := EdgeSource{
		Type:           EdgeSourceTypeForeignKey,
		SourceID:       401,
		RelationType:   domainschema.RelationTypeParentChild,
		ValueSource:    domainproject.RelationValueSourceFromExecution,
		Summary:        "physical foreign key",
		ExternalSource: false,
	}
	logicalSource := EdgeSource{
		Type:           EdgeSourceTypeTableRelation,
		SourceID:       501,
		RelationType:   domainschema.RelationTypeParentChild,
		ValueSource:    domainproject.RelationValueSourceFromExecution,
		Summary:        "schema table relation",
		ExternalSource: false,
	}

	graph.AddEdge(DependencyEdge{FromProjectTableID: 301, ToProjectTableID: 302, Sources: []EdgeSource{foreignKeySource}})
	graph.AddEdge(DependencyEdge{FromProjectTableID: 301, ToProjectTableID: 302, Sources: []EdgeSource{logicalSource}})

	if len(graph.Edges) != 1 {
		t.Fatalf("Edges length = %d, want 1", len(graph.Edges))
	}
	edge := graph.Edges[0]
	if edge.CanonicalKey() != "301->302" {
		t.Fatalf("CanonicalKey = %q, want 301->302", edge.CanonicalKey())
	}
	if len(edge.Sources) != 2 {
		t.Fatalf("Sources length = %d, want 2", len(edge.Sources))
	}
	assertEdgeSource(t, edge.Sources[0], foreignKeySource)
	assertEdgeSource(t, edge.Sources[1], logicalSource)
}

func TestDependencyEdgeCanonicalKeyUsesOrderedEndpoints(t *testing.T) {
	forward := DependencyEdge{FromProjectTableID: 301, ToProjectTableID: 302}
	reverse := DependencyEdge{FromProjectTableID: 302, ToProjectTableID: 301}

	if forward.CanonicalKey() != "301->302" {
		t.Fatalf("forward CanonicalKey = %q, want 301->302", forward.CanonicalKey())
	}
	if reverse.CanonicalKey() != "302->301" {
		t.Fatalf("reverse CanonicalKey = %q, want 302->301", reverse.CanonicalKey())
	}
	if forward.CanonicalKey() == reverse.CanonicalKey() {
		t.Fatal("canonical edge key must preserve dependency direction")
	}
}

func TestEdgeSourceSupportsExternalSourceSummaryWithoutPayload(t *testing.T) {
	source := EdgeSource{
		Type:           EdgeSourceTypeProjectRelation,
		SourceID:       601,
		RelationType:   domainschema.RelationTypeParentChild,
		ValueSource:    domainproject.RelationValueSourceFromDBQuery,
		Summary:        "project relation uses external value source",
		ExternalSource: true,
	}

	if source.Type != EdgeSourceTypeProjectRelation {
		t.Fatalf("Type = %s, want %s", source.Type, EdgeSourceTypeProjectRelation)
	}
	if source.SourceID != 601 {
		t.Fatalf("SourceID = %d, want 601", source.SourceID)
	}
	if source.RelationType != domainschema.RelationTypeParentChild {
		t.Fatalf("RelationType = %s, want %s", source.RelationType, domainschema.RelationTypeParentChild)
	}
	if !source.ExternalSource {
		t.Fatal("expected external source marker")
	}
	if source.ValueSource != domainproject.RelationValueSourceFromDBQuery {
		t.Fatalf("ValueSource = %s, want %s", source.ValueSource, domainproject.RelationValueSourceFromDBQuery)
	}
	assertNoSensitivePlanFragments(t, source.Summary)
}

func TestDependencyGraphDeduplicatesRelationEdgesSortsSourcesAndWarns(t *testing.T) {
	parentProjectTableID := int64(301)
	input := &PlanInput{
		Tables: []PlanTableInput{
			{ProjectTableID: 302, TableID: 202},
			{ProjectTableID: 301, TableID: 201},
		},
		ForeignKeys: []domainschema.ForeignKey{
			{ID: 402, TableID: 202, ReferencedTableID: 201},
			{ID: 401, TableID: 202, ReferencedTableID: 201},
		},
		TableRelations: []domainschema.TableRelation{
			{ID: 501, RelationType: domainschema.RelationTypeParentChild, ParentTableID: 201, ChildTableID: 202},
		},
		ProjectRelations: []domainproject.ProjectTableRelation{
			{
				ID:                   601,
				TableRelationID:      501,
				ParentProjectTableID: &parentProjectTableID,
				ChildProjectTableID:  302,
				RelValueSource:       domainproject.RelationValueSourceFromExecution,
			},
		},
	}

	graph := NewDependencyGraph(input)

	if len(graph.Edges) != 1 {
		t.Fatalf("Edges length = %d, want 1", len(graph.Edges))
	}
	edge := graph.Edges[0]
	if edge.FromProjectTableID != 301 || edge.ToProjectTableID != 302 {
		t.Fatalf("edge endpoints = %d -> %d, want 301 -> 302", edge.FromProjectTableID, edge.ToProjectTableID)
	}
	if len(edge.Sources) != 4 {
		t.Fatalf("Sources length = %d, want 4", len(edge.Sources))
	}
	assertEdgeSource(t, edge.Sources[0], EdgeSource{Type: EdgeSourceTypeForeignKey, SourceID: 401, Summary: "physical foreign key dependency"})
	assertEdgeSource(t, edge.Sources[1], EdgeSource{Type: EdgeSourceTypeForeignKey, SourceID: 402, Summary: "physical foreign key dependency"})
	assertEdgeSource(t, edge.Sources[2], EdgeSource{Type: EdgeSourceTypeTableRelation, SourceID: 501, RelationType: domainschema.RelationTypeParentChild, Summary: "schema table relation dependency"})
	assertEdgeSource(t, edge.Sources[3], EdgeSource{Type: EdgeSourceTypeProjectRelation, SourceID: 601, ValueSource: domainproject.RelationValueSourceFromExecution, Summary: "project relation current execution dependency"})

	warnings := 0
	for _, issue := range graph.Issues {
		if issue.Code == PlanErrorCodeDuplicateEdge && !issue.Blocking {
			warnings++
		}
		assertNoSensitivePlanIssueContent(t, issue)
	}
	if warnings != 3 {
		t.Fatalf("duplicate edge warnings = %d, want 3: %#v", warnings, graph.Issues)
	}
}

func TestDependencyGraphSortsDeduplicatedEdgesByCanonicalKey(t *testing.T) {
	input := &PlanInput{
		Tables: []PlanTableInput{
			{ProjectTableID: 303, TableID: 203},
			{ProjectTableID: 302, TableID: 202},
			{ProjectTableID: 301, TableID: 201},
		},
		ForeignKeys: []domainschema.ForeignKey{
			{ID: 402, TableID: 203, ReferencedTableID: 202},
			{ID: 401, TableID: 202, ReferencedTableID: 201},
		},
	}

	graph := NewDependencyGraph(input)

	if len(graph.Edges) != 2 {
		t.Fatalf("Edges length = %d, want 2", len(graph.Edges))
	}
	if graph.Edges[0].CanonicalKey() != "301->302" {
		t.Fatalf("first edge key = %q, want 301->302", graph.Edges[0].CanonicalKey())
	}
	if graph.Edges[1].CanonicalKey() != "302->303" {
		t.Fatalf("second edge key = %q, want 302->303", graph.Edges[1].CanonicalKey())
	}
}

func TestDependencyGraphSummarizesBlockingErrorsAndWarningsForPrecheck(t *testing.T) {
	graph := DependencyGraph{Issues: []PlanIssue{
		NewPlanIssue(PlanErrorCodeMissingEndpoint, PlanStageRelationMapping, "foreignKeys[0].referencedTableId", "physical foreign key referenced table is outside the dependency graph", true),
		NewPlanIssue(PlanErrorCodeDuplicateEdge, PlanStageGraphBuild, "edges[301->302]", "dependency edge has multiple relation sources", false),
		NewPlanIssue(PlanErrorCodeUnknownValueSource, PlanStageRelationMapping, "projectRelations[0].relValueSource", "project relation value source is not recognized", true),
	}}

	result := SummarizeGraphDiagnostics(graph)

	if result.Passed {
		t.Fatal("expected blocking graph diagnostics to fail precheck")
	}
	if len(result.BlockingErrors) != 2 {
		t.Fatalf("BlockingErrors length = %d, want 2: %#v", len(result.BlockingErrors), result.BlockingErrors)
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("Warnings length = %d, want 1: %#v", len(result.Warnings), result.Warnings)
	}
	assertPlanIssue(t, result.BlockingErrors, "foreignKeys[0].referencedTableId", PlanErrorCodeMissingEndpoint, PlanStageRelationMapping)
	assertPlanIssue(t, result.BlockingErrors, "projectRelations[0].relValueSource", PlanErrorCodeUnknownValueSource, PlanStageRelationMapping)
	if result.Warnings[0].Blocking {
		t.Fatalf("warning was marked blocking: %#v", result.Warnings[0])
	}
	if result.Warnings[0].Code != PlanErrorCodeDuplicateEdge || result.Warnings[0].FieldPath != "edges[301->302]" {
		t.Fatalf("unexpected warning issue: %#v", result.Warnings[0])
	}
}

func TestDependencyGraphDiagnosticsPassWithWarningsOnly(t *testing.T) {
	graph := DependencyGraph{Issues: []PlanIssue{
		NewPlanIssue(PlanErrorCodeDuplicateEdge, PlanStageGraphBuild, "edges[301->302]", "dependency edge has multiple relation sources", false),
	}}

	result := SummarizeGraphDiagnostics(graph)

	if !result.Passed {
		t.Fatalf("expected warnings-only graph diagnostics to pass, got %#v", result.BlockingErrors)
	}
	if len(result.BlockingErrors) != 0 {
		t.Fatalf("BlockingErrors length = %d, want 0", len(result.BlockingErrors))
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("Warnings length = %d, want 1", len(result.Warnings))
	}
	if result.Warnings[0].Blocking {
		t.Fatalf("warning was marked blocking: %#v", result.Warnings[0])
	}
}

func assertDependencyEdge(t *testing.T, got DependencyEdge, fromProjectTableID int64, toProjectTableID int64, source EdgeSource) {
	t.Helper()
	if got.FromProjectTableID != fromProjectTableID || got.ToProjectTableID != toProjectTableID {
		t.Fatalf("edge endpoints = %d -> %d, want %d -> %d", got.FromProjectTableID, got.ToProjectTableID, fromProjectTableID, toProjectTableID)
	}
	if len(got.Sources) != 1 {
		t.Fatalf("Sources length = %d, want 1", len(got.Sources))
	}
	assertEdgeSource(t, got.Sources[0], source)
}

func assertGraphNode(t *testing.T, got GraphNode, want GraphNode) {
	t.Helper()
	if got != want {
		t.Fatalf("GraphNode = %#v, want %#v", got, want)
	}
}

func assertEdgeSource(t *testing.T, got EdgeSource, want EdgeSource) {
	t.Helper()
	if got != want {
		t.Fatalf("EdgeSource = %#v, want %#v", got, want)
	}
}
