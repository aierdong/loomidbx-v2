package execution

import "time"

// ExecutionTask is the execution history master record for one generation run.
type ExecutionTask struct {
	// ID stores the stable execution task identity; new unsaved tasks may use zero and loaded history uses a positive value.
	ID int64 `json:"id"`

	// ProjectID stores the owning Project identity referenced by this execution task.
	ProjectID int64 `json:"projectId"`

	// TaskName stores the user-facing execution task name.
	TaskName string `json:"taskName"`

	// Status stores the current or final task status as a stable string-backed enum.
	Status ExecutionTaskStatus `json:"status"`

	// StartedAt stores when execution started and is required for persisted history.
	StartedAt time.Time `json:"startedAt"`

	// EndedAt stores when execution ended; nil represents a task that has not recorded an end time.
	EndedAt *time.Time `json:"endedAt,omitempty"`

	// CreatedAt stores when the execution task record was created.
	CreatedAt time.Time `json:"createdAt"`
}
