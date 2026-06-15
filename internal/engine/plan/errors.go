package plan

import "strings"

// PlanErrorCode identifies a stable machine-readable dependency planning error category.
type PlanErrorCode string

const (
	// PlanErrorCodeRequired reports that a required dependency planning field is missing.
	PlanErrorCodeRequired PlanErrorCode = "REQUIRED"

	// PlanErrorCodeInvalidReference reports that a dependency planning identity or endpoint reference is invalid.
	PlanErrorCodeInvalidReference PlanErrorCode = "INVALID_REFERENCE"

	// PlanErrorCodeDuplicateNode reports that more than one graph node maps to the same Schema table.
	PlanErrorCodeDuplicateNode PlanErrorCode = "DUPLICATE_NODE"

	// PlanErrorCodeDuplicateEdge reports that multiple relation sources produced the same directed dependency edge.
	PlanErrorCodeDuplicateEdge PlanErrorCode = "DUPLICATE_EDGE"

	// PlanErrorCodeUnknownRelationType reports that a Schema relation type is not recognized.
	PlanErrorCodeUnknownRelationType PlanErrorCode = "UNKNOWN_RELATION_TYPE"

	// PlanErrorCodeUnknownValueSource reports that a Project relation value source is not recognized.
	PlanErrorCodeUnknownValueSource PlanErrorCode = "UNKNOWN_VALUE_SOURCE"

	// PlanErrorCodeMissingEndpoint reports that a dependency edge endpoint cannot be resolved to a graph node.
	PlanErrorCodeMissingEndpoint PlanErrorCode = "MISSING_ENDPOINT"

	// PlanErrorCodeCycleDetected reports that graph dependencies contain a cycle.
	PlanErrorCodeCycleDetected PlanErrorCode = "CYCLE_DETECTED"

	// PlanErrorCodeUnsortableGraph reports that dependency ordering could not include all graph nodes.
	PlanErrorCodeUnsortableGraph PlanErrorCode = "UNSORTABLE_GRAPH"

	// PlanErrorCodeSensitiveValueNotAllowed reports that a candidate public issue field contained sensitive content.
	PlanErrorCodeSensitiveValueNotAllowed PlanErrorCode = "SENSITIVE_VALUE_NOT_ALLOWED"
)

// String returns the stable string representation used by dependency planning boundaries.
func (c PlanErrorCode) String() string {
	return string(c)
}

// PlanStage identifies the dependency planning phase where a safe issue was produced.
type PlanStage string

const (
	// PlanStageInputMapping identifies Project and Schema snapshot input mapping.
	PlanStageInputMapping PlanStage = "INPUT_MAPPING"

	// PlanStageGraphBuild identifies graph node and edge construction.
	PlanStageGraphBuild PlanStage = "GRAPH_BUILD"

	// PlanStageRelationMapping identifies relation-to-edge mapping.
	PlanStageRelationMapping PlanStage = "RELATION_MAPPING"

	// PlanStageTopologicalSort identifies dependency graph sorting.
	PlanStageTopologicalSort PlanStage = "TOPOLOGICAL_SORT"

	// PlanStagePrecheck identifies aggregate dependency planning precheck.
	PlanStagePrecheck PlanStage = "PRECHECK"
)

// String returns the stable string representation used by dependency planning boundaries.
func (s PlanStage) String() string {
	return string(s)
}

// PlanIssue is the only public dependency planning issue shape and contains a safe summary only.
type PlanIssue struct {
	// Code stores the stable machine-readable planning issue category.
	Code PlanErrorCode

	// Stage stores the planning phase where the issue was produced.
	Stage PlanStage

	// FieldPath stores the dependency planning input, graph, relation, or sort path related to the issue.
	FieldPath string

	// SafeMessage stores a user-readable message without credentials, SQL, connection details, or generated data.
	SafeMessage string

	// Blocking reports whether the issue prevents dependency planning from producing a runnable order.
	Blocking bool
}

// PlanPrecheckResult stores blocking errors and warnings produced before dependency planning can continue.
type PlanPrecheckResult struct {
	// Passed reports whether dependency planning precheck found no blocking errors.
	Passed bool

	// BlockingErrors stores safe planning issues that prevent a dependency plan from being used.
	BlockingErrors []PlanIssue

	// Warnings stores safe non-blocking planning issues.
	Warnings []PlanIssue
}

// NewPlanIssue creates a dependency planning issue from already-classified safe error data.
func NewPlanIssue(code PlanErrorCode, stage PlanStage, fieldPath string, safeMessage string, blocking bool) PlanIssue {
	field := strings.TrimSpace(fieldPath)
	if containsSensitivePlanContent(code.String()) || containsSensitivePlanContent(stage.String()) || containsSensitivePlanContent(field) {
		return PlanIssue{
			Code:        PlanErrorCodeSensitiveValueNotAllowed,
			Stage:       safePlanStage(stage),
			FieldPath:   "plan.issue",
			SafeMessage: defaultPlanSafeMessage(PlanErrorCodeSensitiveValueNotAllowed),
			Blocking:    blocking,
		}
	}

	message := strings.TrimSpace(safeMessage)
	if message == "" || containsSensitivePlanContent(message) {
		message = defaultPlanSafeMessage(code)
	}

	return PlanIssue{
		Code:        code,
		Stage:       stage,
		FieldPath:   field,
		SafeMessage: message,
		Blocking:    blocking,
	}
}

func (r *PlanPrecheckResult) addBlockingIssue(fieldPath string, code PlanErrorCode, stage PlanStage, safeMessage string) {
	r.BlockingErrors = append(r.BlockingErrors, NewPlanIssue(code, stage, fieldPath, safeMessage, true))
	r.Passed = false
}

func safePlanStage(stage PlanStage) PlanStage {
	if strings.TrimSpace(stage.String()) == "" || containsSensitivePlanContent(stage.String()) {
		return PlanStagePrecheck
	}
	return stage
}

func defaultPlanSafeMessage(code PlanErrorCode) string {
	switch code {
	case PlanErrorCodeRequired:
		return "required dependency planning field is missing"
	case PlanErrorCodeInvalidReference:
		return "dependency planning reference is invalid"
	case PlanErrorCodeDuplicateNode:
		return "dependency graph node is duplicated"
	case PlanErrorCodeDuplicateEdge:
		return "dependency edge source is duplicated"
	case PlanErrorCodeUnknownRelationType:
		return "dependency relation type is not recognized"
	case PlanErrorCodeUnknownValueSource:
		return "dependency relation value source is not recognized"
	case PlanErrorCodeMissingEndpoint:
		return "dependency edge endpoint is missing"
	case PlanErrorCodeCycleDetected:
		return "dependency graph contains a cycle"
	case PlanErrorCodeUnsortableGraph:
		return "dependency graph cannot be sorted"
	case PlanErrorCodeSensitiveValueNotAllowed:
		return "dependency planning issue contained sensitive content"
	default:
		return "dependency planning issue occurred"
	}
}

func containsSensitivePlanContent(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{
		"password",
		"credential",
		"token",
		"secret",
		"connection string",
		"host=",
		"port=",
		"postgres://",
		"postgresql://",
		"mysql://",
		"sqlserver://",
		"mongodb://",
		"redis://",
		"jdbc:",
		"select ",
		"insert ",
		"update ",
		"delete ",
		"create ",
		"drop ",
		"alter ",
		"user sql",
		"generated data",
		"generateddata",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
