package rule

// DataMappingType identifies the stable logical mapping category for a generator output value.
type DataMappingType string

const (
	// DataMappingTypeText represents generated string text output.
	DataMappingTypeText DataMappingType = "text"

	// DataMappingTypeInteger represents generated integer numeric output.
	DataMappingTypeInteger DataMappingType = "integer"

	// DataMappingTypeFloat represents generated floating-point numeric output.
	DataMappingTypeFloat DataMappingType = "float"

	// DataMappingTypeBoolean represents generated boolean output.
	DataMappingTypeBoolean DataMappingType = "boolean"

	// DataMappingTypeDatetime represents generated date or time output.
	DataMappingTypeDatetime DataMappingType = "datetime"
)

// IsKnown reports whether the data mapping type is one of the stable values owned by this domain model.
func (t DataMappingType) IsKnown() bool {
	switch t {
	case DataMappingTypeText,
		DataMappingTypeInteger,
		DataMappingTypeFloat,
		DataMappingTypeBoolean,
		DataMappingTypeDatetime:
		return true
	default:
		return false
	}
}

// IsUnknown reports whether the data mapping type is outside the stable values owned by this domain model.
func (t DataMappingType) IsUnknown() bool {
	return !t.IsKnown()
}

// String returns the data mapping type as its stable JSON/string contract value.
func (t DataMappingType) String() string {
	return string(t)
}
