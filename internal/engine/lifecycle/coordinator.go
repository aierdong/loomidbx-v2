package lifecycle

import (
	"time"

	domainexecution "github.com/gerdong/loomidbx/internal/domain/execution"
)

// LifecycleCoordinator runs the current lifecycle boundary for one execution task snapshot.
type LifecycleCoordinator struct {
	now   func() time.Time
	ports DownstreamPorts
}

// LifecycleRunResult is the final coordinator result for the current lifecycle boundary.
type LifecycleRunResult struct {
	// Input stores the mapped engine execution input when snapshot validation succeeds.
	Input *ExecutionInput

	// Precheck stores the aggregate precheck result used to decide whether execution can start.
	Precheck PrecheckResult

	// State stores the final lifecycle state after coordinator execution.
	State LifecycleState

	// StartedAt stores the time execution entered RUNNING, when it started.
	StartedAt *time.Time

	// CompletedAt stores the time execution entered COMPLETED, when it completed.
	CompletedAt *time.Time

	// Control stores the downstream-observable runtime control token.
	Control ControlToken

	// Failure stores the safe failure summary when the lifecycle failed.
	Failure *LifecycleError

	// Transitions stores accepted and rejected lifecycle transition records.
	Transitions []TransitionRecord
}

// NewLifecycleCoordinator creates a coordinator with an injectable clock for deterministic lifecycle tests.
func NewLifecycleCoordinator(now func() time.Time) LifecycleCoordinator {
	return NewLifecycleCoordinatorWithPorts(now, NoopDownstreamPorts())
}

// NewLifecycleCoordinatorWithPorts creates a coordinator with replaceable downstream lifecycle seams.
func NewLifecycleCoordinatorWithPorts(now func() time.Time, ports DownstreamPorts) LifecycleCoordinator {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return LifecycleCoordinator{now: now, ports: normalizeDownstreamPorts(ports)}
}

// Run maps a generation job snapshot, prechecks it, starts execution, and completes the current boundary.
func (c LifecycleCoordinator) Run(job domainexecution.GenerationJob) LifecycleRunResult {
	machine := NewLifecycle()
	control := machine.ControlToken()
	input, basePrecheck := MapExecutionInputFromGenerationJob(job)

	machine.TransitionTo(LifecycleStatePrechecking, c.now())
	precheck := AggregatePrecheck(input, machine, basePrecheck, c.downstreamPrecheck(input, control))
	if !precheck.Passed {
		failure := firstPrecheckFailure(precheck)
		machine.TransitionTo(LifecycleStateFailed, c.now())
		return LifecycleRunResult{
			Input:       input,
			Precheck:    precheck,
			State:       machine.State(),
			Control:     control,
			Failure:     failure,
			Transitions: machine.TransitionRecords(),
		}
	}

	machine.TransitionTo(LifecycleStateReady, c.now())
	startedAt := c.now()
	machine.TransitionTo(LifecycleStateRunning, startedAt)

	context := DownstreamContext{Input: input, Control: control}
	plan := c.ports.Planner.Plan(context)
	if plan.Status == DownstreamStageFailed {
		return c.failFromDownstream(machine, input, precheck, control, startedAt, LifecycleStagePlanner, "planner", plan.Failure)
	}

	context.PlanArtifact = plan.Artifact
	generation := c.ports.Generation.Generate(context)
	if generation.Status == DownstreamStageFailed {
		return c.failFromDownstream(machine, input, precheck, control, startedAt, LifecycleStageGeneration, "generation", generation.Failure)
	}

	context.GenerationArtifact = generation.Artifact
	result := c.ports.Result.Summarize(context)
	if result.Status == DownstreamStageFailed {
		return c.failFromDownstream(machine, input, precheck, control, startedAt, LifecycleStageResult, "result", result.Failure)
	}

	completedAt := c.now()
	machine.TransitionTo(LifecycleStateCompleted, completedAt)

	return LifecycleRunResult{
		Input:       input,
		Precheck:    precheck,
		State:       machine.State(),
		StartedAt:   &startedAt,
		CompletedAt: &completedAt,
		Control:     control,
		Transitions: machine.TransitionRecords(),
	}
}

func (c LifecycleCoordinator) failFromDownstream(machine *Lifecycle, input *ExecutionInput, precheck PrecheckResult, control ControlToken, startedAt time.Time, stage LifecycleStage, fieldPath string, downstreamFailure *LifecycleError) LifecycleRunResult {
	failure := MapDownstreamStageFailure(stage, fieldPath, downstreamFailure)
	machine.TransitionTo(LifecycleStateFailed, c.now())
	return LifecycleRunResult{
		Input:       input,
		Precheck:    precheck,
		State:       machine.State(),
		StartedAt:   &startedAt,
		Control:     control,
		Failure:     &failure,
		Transitions: machine.TransitionRecords(),
	}
}

func firstPrecheckFailure(precheck PrecheckResult) *LifecycleError {
	if len(precheck.BlockingErrors) == 0 {
		return nil
	}
	failure := precheck.BlockingErrors[0]
	failure.Stage = LifecycleStagePrecheck
	return &failure
}

func normalizeDownstreamPorts(ports DownstreamPorts) DownstreamPorts {
	defaults := NoopDownstreamPorts()
	if ports.Precheck == nil {
		ports.Precheck = defaults.Precheck
	}
	if ports.Planner == nil {
		ports.Planner = defaults.Planner
	}
	if ports.Generation == nil {
		ports.Generation = defaults.Generation
	}
	if ports.Result == nil {
		ports.Result = defaults.Result
	}
	return ports
}

func (c LifecycleCoordinator) downstreamPrecheck(input *ExecutionInput, control ControlToken) PrecheckResult {
	if c.ports.Precheck == nil || input == nil {
		return PrecheckResult{Passed: true}
	}
	return c.ports.Precheck.Precheck(DownstreamContext{Input: input, Control: control})
}
