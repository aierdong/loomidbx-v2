// Package project contains pure domain models for Project task organization.
package project

import "time"

// Project is the aggregate root for a reusable generation task configuration.
type Project struct {
	// ID stores the stable Project identity; drafts use zero and persisted Projects use a positive value.
	ID int64 `json:"id"`

	// ConnectionID stores the target database connection identity referenced by this Project.
	ConnectionID int64 `json:"connectionId"`

	// Name stores the user-facing Project name.
	Name string `json:"name"`

	// Description stores optional user-facing notes about the Project.
	Description string `json:"description"`

	// CreatedAt stores when the persisted Project was created; drafts may keep the zero time.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt stores when the persisted Project was last changed; drafts may keep the zero time.
	UpdatedAt time.Time `json:"updatedAt"`
}
