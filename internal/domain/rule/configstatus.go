package rule

// ConfigStatus identifies the stable lifecycle state for a field generator configuration.
type ConfigStatus string

const (
	// ConfigStatusActive marks a generator configuration that can be used by later generation flows.
	ConfigStatusActive ConfigStatus = "ACTIVE"

	// ConfigStatusNeedsReview marks a generator configuration that requires human review after schema changes.
	ConfigStatusNeedsReview ConfigStatus = "NEEDS_REVIEW"
)

// IsKnown reports whether the config status is one of the stable values owned by this domain model.
func (s ConfigStatus) IsKnown() bool {
	switch s {
	case ConfigStatusActive,
		ConfigStatusNeedsReview:
		return true
	default:
		return false
	}
}

// IsUnknown reports whether the config status is outside the stable values owned by this domain model.
func (s ConfigStatus) IsUnknown() bool {
	return !s.IsKnown()
}

// String returns the config status as its stable JSON/string contract value.
func (s ConfigStatus) String() string {
	return string(s)
}
