package lifecycle

// AggregatePrecheck merges input validation, lifecycle state prerequisites, and downstream precheck results.
// It returns a safe summary where Passed is true only when no blocking errors are present.
func AggregatePrecheck(input *ExecutionInput, machine *Lifecycle, base PrecheckResult, downstreamResults ...PrecheckResult) PrecheckResult {
	result := PrecheckResult{Passed: true}
	result.append(base)

	if input == nil {
		result.addBlockingIssue("input", LifecycleErrorCodeRequired, "execution input is required before execution can start")
	}
	if machine == nil || !canAggregatePrecheck(machine.State()) {
		state := ""
		if machine != nil {
			state = machine.State().String()
		}
		conflict := MapStateConflictError(LifecycleStagePrecheck, "state", state, LifecycleStatePrechecking.String())
		result.BlockingErrors = append(result.BlockingErrors, conflict)
	}

	for _, downstream := range downstreamResults {
		result.append(downstream)
	}

	result.Passed = len(result.BlockingErrors) == 0
	return result
}

func canAggregatePrecheck(state LifecycleState) bool {
	return state == LifecycleStateInitialized || state == LifecycleStatePrechecking
}

func (r *PrecheckResult) append(other PrecheckResult) {
	if len(other.BlockingErrors) > 0 {
		r.BlockingErrors = append(r.BlockingErrors, other.BlockingErrors...)
	}
	if len(other.Warnings) > 0 {
		r.Warnings = append(r.Warnings, other.Warnings...)
	}
}
