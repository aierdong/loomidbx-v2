package rule

// ConfigStatus identifies the stable lifecycle state for a field generator configuration.
type ConfigStatus string

const (
	// ConfigStatusActive marks a generator configuration that can be used by later generation flows.
	ConfigStatusActive ConfigStatus = "ACTIVE"

	// ConfigStatusNeedsReview marks a generator configuration that requires human review after schema changes.
	ConfigStatusNeedsReview ConfigStatus = "NEEDS_REVIEW"
)
