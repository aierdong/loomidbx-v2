package rule

import "encoding/json"

// GeneratorParams stores a generator-specific JSON payload without binding to a concrete generator schema.
type GeneratorParams struct {
	// Raw stores the nullable raw JSON payload supplied for a generator configuration.
	Raw json.RawMessage `json:"-"`
}

// MarshalJSON returns the generator params as the raw JSON payload, using null when no payload is present.
func (p GeneratorParams) MarshalJSON() ([]byte, error) {
	if len(p.Raw) == 0 || string(p.Raw) == "null" {
		return []byte("null"), nil
	}

	return json.RawMessage(p.Raw).MarshalJSON()
}

// UnmarshalJSON stores the raw generator params JSON payload and normalizes null to no payload.
func (p *GeneratorParams) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		p.Raw = nil
		return nil
	}

	p.Raw = append(p.Raw[:0], data...)
	return nil
}
