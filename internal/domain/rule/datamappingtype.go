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
