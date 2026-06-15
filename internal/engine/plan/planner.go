package plan

import (
	domainproject "github.com/gerdong/loomidbx/internal/domain/project"
	domainschema "github.com/gerdong/loomidbx/internal/domain/schema"
	"github.com/gerdong/loomidbx/internal/engine/lifecycle"
)

// ExecutionPlan stores the table-level dependency order artifact produced by the dependency planner.
type ExecutionPlan struct {
	// ProjectID stores the Project identity represented by this dependency plan.
	ProjectID int64

	// OrderedTables stores every graph node in deterministic dependency execution order.
	OrderedTables []PlannedTable

	// Edges stores deduplicated dependency edges used to produce the order.
	Edges []DependencyEdge

	// Warnings stores safe non-blocking dependency planning issues.
	Warnings []PlanIssue
}

// PlanResult stores either a successful execution plan or safe blocking errors and warnings.
type PlanResult struct {
	// Passed reports whether dependency planning produced a successful execution plan.
	Passed bool

	// Plan stores the successful dependency execution plan when Passed is true.
	Plan *ExecutionPlan

	// BlockingErrors stores safe planning issues that prevent execution from starting.
	BlockingErrors []PlanIssue

	// Warnings stores safe non-blocking planning issues.
	Warnings []PlanIssue
}

// DependencyPlanner adapts dependency planning to lifecycle precheck and planner seams.
type DependencyPlanner struct {
	projectSnapshot  domainproject.Project
	tables           []domainproject.ProjectTable
	foreignKeys      []domainschema.ForeignKey
	tableRelations   []domainschema.TableRelation
	projectRelations []domainproject.ProjectTableRelation
}

// NewDependencyPlanner creates a lifecycle-compatible dependency planner from Project and Schema snapshots.
func NewDependencyPlanner(projectSnapshot domainproject.Project, tables []domainproject.ProjectTable, foreignKeys []domainschema.ForeignKey, tableRelations []domainschema.TableRelation, projectRelations []domainproject.ProjectTableRelation) DependencyPlanner {
	return DependencyPlanner{
		projectSnapshot:  projectSnapshot,
		tables:           tables,
		foreignKeys:      foreignKeys,
		tableRelations:   tableRelations,
		projectRelations: projectRelations,
	}
}

// PlanDependencies builds a dependency graph, sorts it, and returns either a plan artifact or safe issues.
func PlanDependencies(projectSnapshot domainproject.Project, tables []domainproject.ProjectTable, foreignKeys []domainschema.ForeignKey, tableRelations []domainschema.TableRelation, projectRelations []domainproject.ProjectTableRelation) PlanResult {
	input, inputPrecheck := BuildPlanInput(projectSnapshot, tables, foreignKeys, tableRelations, projectRelations)
	if !inputPrecheck.Passed {
		return failedPlanResult(inputPrecheck.BlockingErrors, inputPrecheck.Warnings)
	}

	graph := NewDependencyGraph(input)
	graphPrecheck := SummarizeGraphDiagnostics(graph)
	if !graphPrecheck.Passed {
		return failedPlanResult(graphPrecheck.BlockingErrors, graphPrecheck.Warnings)
	}

	orderedTables, sortIssues := SortDependencyGraph(graph)
	if len(sortIssues) > 0 {
		warnings := append([]PlanIssue{}, graphPrecheck.Warnings...)
		return failedPlanResult(sortIssues, warnings)
	}

	warnings := append([]PlanIssue{}, graphPrecheck.Warnings...)
	plan := &ExecutionPlan{
		ProjectID:     input.ProjectID,
		OrderedTables: append([]PlannedTable{}, orderedTables...),
		Edges:         cloneDependencyEdges(graph.Edges),
		Warnings:      append([]PlanIssue{}, warnings...),
	}
	return PlanResult{Passed: true, Plan: plan, Warnings: warnings}
}

// Precheck returns dependency planning issues in lifecycle precheck result shape.
func (p DependencyPlanner) Precheck(context lifecycle.DownstreamContext) lifecycle.PrecheckResult {
	result := p.plan()
	return lifecycle.PrecheckResult{
		Passed:         result.Passed,
		BlockingErrors: mapPlanIssuesToLifecycle(result.BlockingErrors, lifecycle.LifecycleStagePrecheck),
		Warnings:       mapPlanIssuesToLifecycle(result.Warnings, lifecycle.LifecycleStagePrecheck),
	}
}

// Plan returns the dependency execution plan as the lifecycle planner artifact.
func (p DependencyPlanner) Plan(context lifecycle.DownstreamContext) lifecycle.DownstreamStageResult {
	result := p.plan()
	if result.Passed {
		return lifecycle.NewDownstreamStageSuccess(result.Plan)
	}
	failure := firstLifecyclePlanFailure(result.BlockingErrors)
	return lifecycle.NewDownstreamStageFailure(failure)
}

func (p DependencyPlanner) plan() PlanResult {
	return PlanDependencies(p.projectSnapshot, p.tables, p.foreignKeys, p.tableRelations, p.projectRelations)
}

func failedPlanResult(blockingErrors []PlanIssue, warnings []PlanIssue) PlanResult {
	return PlanResult{
		Passed:         false,
		BlockingErrors: append([]PlanIssue{}, blockingErrors...),
		Warnings:       append([]PlanIssue{}, warnings...),
	}
}

func firstLifecyclePlanFailure(issues []PlanIssue) lifecycle.LifecycleError {
	if len(issues) == 0 {
		return lifecycle.NewLifecycleError(
			lifecycle.LifecycleErrorCodeDownstreamFailure,
			lifecycle.LifecycleStagePlanner,
			"planner",
			"dependency planner failed",
		)
	}
	return mapPlanIssueToLifecycle(issues[0], lifecycle.LifecycleStagePlanner)
}

func mapPlanIssuesToLifecycle(issues []PlanIssue, stage lifecycle.LifecycleStage) []lifecycle.PrecheckIssue {
	mapped := make([]lifecycle.PrecheckIssue, 0, len(issues))
	for _, issue := range issues {
		mapped = append(mapped, mapPlanIssueToLifecycle(issue, stage))
	}
	return mapped
}

func mapPlanIssueToLifecycle(issue PlanIssue, stage lifecycle.LifecycleStage) lifecycle.LifecycleError {
	return lifecycle.NewLifecycleError(
		mapPlanCodeToLifecycle(issue.Code),
		stage,
		issue.FieldPath,
		issue.SafeMessage,
	)
}

func mapPlanCodeToLifecycle(code PlanErrorCode) lifecycle.LifecycleErrorCode {
	switch code {
	case PlanErrorCodeRequired:
		return lifecycle.LifecycleErrorCodeRequired
	case PlanErrorCodeInvalidReference, PlanErrorCodeDuplicateNode, PlanErrorCodeMissingEndpoint:
		return lifecycle.LifecycleErrorCodeInvalidReference
	case PlanErrorCodeSensitiveValueNotAllowed:
		return lifecycle.LifecycleErrorCodeSensitiveValueNotAllowed
	default:
		return lifecycle.LifecycleErrorCodeDownstreamFailure
	}
}

func cloneDependencyEdges(edges []DependencyEdge) []DependencyEdge {
	cloned := make([]DependencyEdge, 0, len(edges))
	for _, edge := range edges {
		edge.Sources = append([]EdgeSource{}, edge.Sources...)
		cloned = append(cloned, edge)
	}
	return cloned
}
