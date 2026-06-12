package lifecycle

// DownstreamStageStatus identifies whether a downstream seam completed or failed.
type DownstreamStageStatus string

const (
	// DownstreamStageSucceeded means the downstream seam finished without a lifecycle failure.
	DownstreamStageSucceeded DownstreamStageStatus = "SUCCEEDED"

	// DownstreamStageFailed means the downstream seam returned a safe lifecycle failure summary.
	DownstreamStageFailed DownstreamStageStatus = "FAILED"
)

// DownstreamContext carries only the lifecycle-owned context needed by downstream seams.
type DownstreamContext struct {
	// Input stores the mapped execution input without dependency graphs, row-count plans, batches, or final results.
	Input *ExecutionInput

	// Control exposes cancellation intent to replaceable downstream seams.
	Control ControlToken

	// PlanArtifact stores an opaque successful planner artifact for later seams.
	PlanArtifact any

	// GenerationArtifact stores an opaque successful generation artifact for the result seam.
	GenerationArtifact any
}

// DownstreamStageResult is the minimal success or failure result shape returned by downstream seams.
type DownstreamStageResult struct {
	// Status reports whether the stage succeeded or failed.
	Status DownstreamStageStatus

	// Failure stores a safe lifecycle error when Status is FAILED.
	Failure *LifecycleError

	// Artifact stores an opaque stage artifact for the next seam without lifecycle interpretation.
	Artifact any
}

// DownstreamPrechecker supplies optional downstream precheck results for aggregate lifecycle precheck.
type DownstreamPrechecker interface {
	// Precheck returns blocking errors or warnings that lifecycle precheck can aggregate safely.
	Precheck(context DownstreamContext) PrecheckResult
}

// PlannerPort defines the replaceable planning seam without dependency graph or row-count planning algorithms.
type PlannerPort interface {
	// Plan returns only a stage success artifact or a safe stage failure summary.
	Plan(context DownstreamContext) DownstreamStageResult
}

// GenerationPort defines the replaceable generation seam without generator registry or batch-loop algorithms.
type GenerationPort interface {
	// Generate returns only a stage success artifact or a safe stage failure summary.
	Generate(context DownstreamContext) DownstreamStageResult
}

// ResultPort defines the replaceable result-summary seam without result aggregation, writer adapters, or real writes.
type ResultPort interface {
	// Summarize returns only a stage success artifact or a safe stage failure summary.
	Summarize(context DownstreamContext) DownstreamStageResult
}

// DownstreamPorts groups replaceable lifecycle seams for planning, generation, and result summary.
type DownstreamPorts struct {
	// Precheck optionally returns downstream precheck issues to aggregate before the lifecycle starts.
	Precheck DownstreamPrechecker

	// Planner runs the minimal planning seam.
	Planner PlannerPort

	// Generation runs the minimal generation seam.
	Generation GenerationPort

	// Result runs the minimal result-summary seam.
	Result ResultPort
}

// NewDownstreamStageSuccess creates a successful downstream stage result with an opaque artifact.
func NewDownstreamStageSuccess(artifact any) DownstreamStageResult {
	return DownstreamStageResult{Status: DownstreamStageSucceeded, Artifact: artifact}
}

// NewDownstreamStageFailure creates a failed downstream stage result with a safe lifecycle error summary.
func NewDownstreamStageFailure(failure LifecycleError) DownstreamStageResult {
	return DownstreamStageResult{Status: DownstreamStageFailed, Failure: &failure}
}

// NoopDownstreamPorts returns replaceable no-op seams that only report stage success.
func NoopDownstreamPorts() DownstreamPorts {
	noop := noopDownstreamPort{}
	return DownstreamPorts{
		Precheck:   noop,
		Planner:    noop,
		Generation: noop,
		Result:     noop,
	}
}

type noopDownstreamPort struct{}

func (noopDownstreamPort) Precheck(context DownstreamContext) PrecheckResult {
	return PrecheckResult{Passed: true}
}

func (noopDownstreamPort) Plan(context DownstreamContext) DownstreamStageResult {
	return NewDownstreamStageSuccess(nil)
}

func (noopDownstreamPort) Generate(context DownstreamContext) DownstreamStageResult {
	return NewDownstreamStageSuccess(nil)
}

func (noopDownstreamPort) Summarize(context DownstreamContext) DownstreamStageResult {
	return NewDownstreamStageSuccess(nil)
}
