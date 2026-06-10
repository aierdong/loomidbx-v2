package rule

import "encoding/json"

// GeneratorParams stores a generator-specific JSON payload without binding to a concrete generator schema.
type GeneratorParams struct {
	// Raw stores the nullable raw JSON payload supplied for a generator configuration.
	Raw json.RawMessage `json:"-"`
}
