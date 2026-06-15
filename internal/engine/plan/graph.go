package plan

import (
	"sort"
	"strconv"

	domainproject "github.com/gerdong/loomidbx/internal/domain/project"
	domainschema "github.com/gerdong/loomidbx/internal/domain/schema"
)

// EdgeSourceType identifies the source model that produced a dependency edge.
type EdgeSourceType string

const (
	// EdgeSourceTypeForeignKey identifies a physical Schema foreign key source.
	EdgeSourceTypeForeignKey EdgeSourceType = "FOREIGN_KEY"

	// EdgeSourceTypeTableRelation identifies a Schema-level logical table relation source.
	EdgeSourceTypeTableRelation EdgeSourceType = "TABLE_RELATION"

	// EdgeSourceTypeProjectRelation identifies a Project-level relation configuration source.
	EdgeSourceTypeProjectRelation EdgeSourceType = "PROJECT_RELATION"
)

// String returns the stable string representation used by dependency graph boundaries.
func (t EdgeSourceType) String() string {
	return string(t)
}

// GraphNode stores one ProjectTable node in the dependency graph.
type GraphNode struct {
	// ProjectTableID stores the ProjectTable identity used by dependency edges and final ordering.
	ProjectTableID int64

	// TableID stores the Schema table identity represented by this graph node.
	TableID int64

	// StableOrder stores the persisted order used as a deterministic tie-breaker.
	StableOrder int
}

// EdgeSource stores a safe summary of the source that produced a dependency edge.
type EdgeSource struct {
	// Type stores the source model category for this edge source.
	Type EdgeSourceType

	// SourceID stores the identity of the source model when available.
	SourceID int64

	// RelationType stores the canonical Schema relation type associated with the source.
	RelationType domainschema.RelationType

	// ValueSource stores the canonical Project relation value source associated with the source.
	ValueSource domainproject.RelationValueSource

	// Summary stores a safe external-source or relation summary without SQL, connection details, or generated data.
	Summary string

	// ExternalSource reports whether the source depends on values outside the current execution.
	ExternalSource bool
}

// DependencyEdge stores a directed dependency from an upstream ProjectTable node to a downstream ProjectTable node.
type DependencyEdge struct {
	// FromProjectTableID stores the upstream ProjectTable node identity.
	FromProjectTableID int64

	// ToProjectTableID stores the downstream ProjectTable node identity.
	ToProjectTableID int64

	// Sources stores the safe source summaries that produced this directed edge.
	Sources []EdgeSource
}

// CanonicalKey returns the stable directed identity used to deduplicate dependency edges.
func (e DependencyEdge) CanonicalKey() string {
	return strconv.FormatInt(e.FromProjectTableID, 10) + "->" + strconv.FormatInt(e.ToProjectTableID, 10)
}

// DependencyGraph stores dependency planning nodes, deduplicated directed edges, and safe graph issues.
type DependencyGraph struct {
	// Nodes stores ProjectTable graph nodes in mapped input order.
	Nodes []GraphNode

	// Edges stores deduplicated directed dependency edges.
	Edges []DependencyEdge

	// ExternalSources stores safe relation source summaries that do not create execution dependency edges.
	ExternalSources []EdgeSource

	// Issues stores safe graph construction issues.
	Issues []PlanIssue
}

// NewDependencyGraph maps validated plan input into dependency graph nodes and relation-derived edges.
func NewDependencyGraph(input *PlanInput) DependencyGraph {
	if input == nil {
		return DependencyGraph{}
	}

	graph := DependencyGraph{Nodes: make([]GraphNode, 0, len(input.Tables))}
	nodesByTableID := make(map[int64]GraphNode, len(input.Tables))
	for _, table := range input.Tables {
		node := GraphNode{
			ProjectTableID: table.ProjectTableID,
			TableID:        table.TableID,
			StableOrder:    table.ExistingExecutionOrder,
		}
		graph.Nodes = append(graph.Nodes, node)
		nodesByTableID[node.TableID] = node
	}

	graph.addForeignKeyEdges(input.ForeignKeys, nodesByTableID)
	graph.addTableRelationEdges(input.TableRelations, nodesByTableID)
	graph.addProjectRelationEdges(input.ProjectRelations)
	graph.sortEdges()

	return graph
}

func (g *DependencyGraph) addForeignKeyEdges(foreignKeys []domainschema.ForeignKey, nodesByTableID map[int64]GraphNode) {
	for index, foreignKey := range foreignKeys {
		prefix := "foreignKeys[" + strconv.Itoa(index) + "]"
		parentNode, parentExists := nodesByTableID[foreignKey.ReferencedTableID]
		childNode, childExists := nodesByTableID[foreignKey.TableID]
		if !parentExists {
			g.addBlockingIssue(prefix+".referencedTableId", PlanErrorCodeMissingEndpoint, "physical foreign key referenced table is outside the dependency graph")
		}
		if !childExists {
			g.addBlockingIssue(prefix+".tableId", PlanErrorCodeMissingEndpoint, "physical foreign key table is outside the dependency graph")
		}
		if !parentExists || !childExists {
			continue
		}
		g.AddEdge(DependencyEdge{
			FromProjectTableID: parentNode.ProjectTableID,
			ToProjectTableID:   childNode.ProjectTableID,
			Sources: []EdgeSource{{
				Type:     EdgeSourceTypeForeignKey,
				SourceID: foreignKey.ID,
				Summary:  "physical foreign key dependency",
			}},
		})
	}
}

func (g *DependencyGraph) addTableRelationEdges(tableRelations []domainschema.TableRelation, nodesByTableID map[int64]GraphNode) {
	for index, relation := range tableRelations {
		prefix := "tableRelations[" + strconv.Itoa(index) + "]"
		if relation.RelationType.IsUnknown() {
			g.addBlockingIssue(prefix+".relationType", PlanErrorCodeUnknownRelationType, "schema table relation type is not recognized")
			continue
		}
		parentNode, parentExists := nodesByTableID[relation.ParentTableID]
		childNode, childExists := nodesByTableID[relation.ChildTableID]
		if !parentExists {
			g.addBlockingIssue(prefix+".parentTableId", PlanErrorCodeMissingEndpoint, "schema table relation parent table is outside the dependency graph")
		}
		if !childExists {
			g.addBlockingIssue(prefix+".childTableId", PlanErrorCodeMissingEndpoint, "schema table relation child table is outside the dependency graph")
		}
		if !parentExists || !childExists {
			continue
		}
		g.AddEdge(DependencyEdge{
			FromProjectTableID: parentNode.ProjectTableID,
			ToProjectTableID:   childNode.ProjectTableID,
			Sources: []EdgeSource{{
				Type:         EdgeSourceTypeTableRelation,
				SourceID:     relation.ID,
				RelationType: relation.RelationType,
				Summary:      "schema table relation dependency",
			}},
		})
	}
}

func (g *DependencyGraph) addProjectRelationEdges(projectRelations []domainproject.ProjectTableRelation) {
	nodesByProjectTableID := make(map[int64]GraphNode, len(g.Nodes))
	for _, node := range g.Nodes {
		nodesByProjectTableID[node.ProjectTableID] = node
	}

	for index, relation := range projectRelations {
		prefix := "projectRelations[" + strconv.Itoa(index) + "]"
		if !relation.RelValueSource.IsKnown() {
			g.addBlockingIssue(prefix+".relValueSource", PlanErrorCodeUnknownValueSource, "project relation value source is not recognized")
			continue
		}

		_, childExists := nodesByProjectTableID[relation.ChildProjectTableID]
		if !childExists {
			g.addBlockingIssue(prefix+".childProjectTableId", PlanErrorCodeMissingEndpoint, "project relation child table is outside the dependency graph")
		}

		switch relation.RelValueSource {
		case domainproject.RelationValueSourceFromExecution:
			g.addCurrentExecutionProjectRelationEdge(prefix, relation, nodesByProjectTableID, childExists, "project relation current execution dependency", false)
		case domainproject.RelationValueSourceFromDBQuery:
			if childExists {
				g.ExternalSources = append(g.ExternalSources, EdgeSource{
					Type:           EdgeSourceTypeProjectRelation,
					SourceID:       relation.ID,
					ValueSource:    relation.RelValueSource,
					Summary:        "project relation external value source",
					ExternalSource: true,
				})
			}
		case domainproject.RelationValueSourceMerged:
			g.addCurrentExecutionProjectRelationEdge(prefix, relation, nodesByProjectTableID, childExists, "project relation merged value dependency", true)
		}
	}
}

func (g *DependencyGraph) addCurrentExecutionProjectRelationEdge(prefix string, relation domainproject.ProjectTableRelation, nodesByProjectTableID map[int64]GraphNode, childExists bool, summary string, externalSource bool) {
	if relation.ParentProjectTableID == nil {
		g.addBlockingIssue(prefix+".parentProjectTableId", PlanErrorCodeMissingEndpoint, "project relation parent table is required for current execution values")
		return
	}

	parentNode, parentExists := nodesByProjectTableID[*relation.ParentProjectTableID]
	if !parentExists {
		g.addBlockingIssue(prefix+".parentProjectTableId", PlanErrorCodeMissingEndpoint, "project relation parent table is outside the dependency graph")
	}
	if !childExists || !parentExists {
		return
	}

	g.AddEdge(DependencyEdge{
		FromProjectTableID: parentNode.ProjectTableID,
		ToProjectTableID:   relation.ChildProjectTableID,
		Sources: []EdgeSource{{
			Type:           EdgeSourceTypeProjectRelation,
			SourceID:       relation.ID,
			ValueSource:    relation.RelValueSource,
			Summary:        summary,
			ExternalSource: externalSource,
		}},
	})
}

func (g *DependencyGraph) addBlockingIssue(fieldPath string, code PlanErrorCode, safeMessage string) {
	g.Issues = append(g.Issues, NewPlanIssue(code, PlanStageRelationMapping, fieldPath, safeMessage, true))
}

func (g *DependencyGraph) addWarningIssue(fieldPath string, code PlanErrorCode, safeMessage string) {
	g.Issues = append(g.Issues, NewPlanIssue(code, PlanStageGraphBuild, fieldPath, safeMessage, false))
}

// SummarizeGraphDiagnostics separates graph issues into blocking errors and warnings for precheck consumers.
func SummarizeGraphDiagnostics(graph DependencyGraph) PlanPrecheckResult {
	result := PlanPrecheckResult{Passed: true}
	for _, issue := range graph.Issues {
		if issue.Blocking {
			result.BlockingErrors = append(result.BlockingErrors, issue)
			result.Passed = false
			continue
		}
		result.Warnings = append(result.Warnings, issue)
	}
	return result
}

// AddEdge adds a dependency edge or merges sources into an existing edge with the same canonical key.
func (g *DependencyGraph) AddEdge(edge DependencyEdge) {
	key := edge.CanonicalKey()
	for index := range g.Edges {
		if g.Edges[index].CanonicalKey() == key {
			g.Edges[index].Sources = append(g.Edges[index].Sources, edge.Sources...)
			sortEdgeSources(g.Edges[index].Sources)
			g.addWarningIssue("edges["+key+"]", PlanErrorCodeDuplicateEdge, "dependency edge has multiple relation sources")
			return
		}
	}
	sortEdgeSources(edge.Sources)
	g.Edges = append(g.Edges, edge)
}

func (g *DependencyGraph) sortEdges() {
	for index := range g.Edges {
		sortEdgeSources(g.Edges[index].Sources)
	}
	sort.SliceStable(g.Edges, func(i, j int) bool {
		return g.Edges[i].CanonicalKey() < g.Edges[j].CanonicalKey()
	})
}

func sortEdgeSources(sources []EdgeSource) {
	sort.SliceStable(sources, func(i, j int) bool {
		return compareEdgeSources(sources[i], sources[j]) < 0
	})
}

func compareEdgeSources(left EdgeSource, right EdgeSource) int {
	if left.Type != right.Type {
		return edgeSourceTypeRank(left.Type) - edgeSourceTypeRank(right.Type)
	}
	if left.SourceID != right.SourceID {
		if left.SourceID < right.SourceID {
			return -1
		}
		return 1
	}
	if left.RelationType != right.RelationType {
		if left.RelationType.String() < right.RelationType.String() {
			return -1
		}
		return 1
	}
	if left.ValueSource != right.ValueSource {
		if string(left.ValueSource) < string(right.ValueSource) {
			return -1
		}
		return 1
	}
	if left.Summary != right.Summary {
		if left.Summary < right.Summary {
			return -1
		}
		return 1
	}
	if left.ExternalSource == right.ExternalSource {
		return 0
	}
	if !left.ExternalSource {
		return -1
	}
	return 1
}

func edgeSourceTypeRank(sourceType EdgeSourceType) int {
	switch sourceType {
	case EdgeSourceTypeForeignKey:
		return 1
	case EdgeSourceTypeTableRelation:
		return 2
	case EdgeSourceTypeProjectRelation:
		return 3
	default:
		return 4
	}
}
