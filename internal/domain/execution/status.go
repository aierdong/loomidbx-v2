package execution

// ExecutionTaskStatus identifies the stable task-level execution status.
type ExecutionTaskStatus string

const (
	// ExecutionTaskStatusRunning means the execution task is currently running.
	ExecutionTaskStatusRunning ExecutionTaskStatus = "RUNNING"

	// ExecutionTaskStatusSuccess means every table in the execution task completed successfully.
	ExecutionTaskStatusSuccess ExecutionTaskStatus = "SUCCESS"

	// ExecutionTaskStatusPartialFailed means some tables failed while successfully written data is retained.
	ExecutionTaskStatusPartialFailed ExecutionTaskStatus = "PARTIAL_FAILED"

	// ExecutionTaskStatusFailed means the task failed at task level or all table executions failed.
	ExecutionTaskStatusFailed ExecutionTaskStatus = "FAILED"
)

// ExecutionTableStatus identifies the stable table-level execution status.
type ExecutionTableStatus string

const (
	// ExecutionTableStatusPending means the table waits for prerequisite tables to complete.
	ExecutionTableStatusPending ExecutionTableStatus = "PENDING"

	// ExecutionTableStatusRunning means the table execution is currently running.
	ExecutionTableStatusRunning ExecutionTableStatus = "RUNNING"

	// ExecutionTableStatusSuccess means rows were written successfully for the table.
	ExecutionTableStatusSuccess ExecutionTableStatus = "SUCCESS"

	// ExecutionTableStatusFailed means writing rows for the table failed.
	ExecutionTableStatusFailed ExecutionTableStatus = "FAILED"

	// ExecutionTableStatusSkipped means the table was skipped because a prerequisite table failed.
	ExecutionTableStatusSkipped ExecutionTableStatus = "SKIPPED"
)
