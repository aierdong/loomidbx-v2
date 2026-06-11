// Package execution contains pure domain models for generation jobs and execution history.
package execution

// GenerationJob describes one generation execution aggregate without owning lifecycle algorithms.
type GenerationJob struct {
	// Task stores the execution task master record for this generation job.
	Task ExecutionTask `json:"task"`

	// TableResults stores table-level execution result records associated with the task.
	TableResults []ExecutionTableResult `json:"tableResults"`
}
