package schema

// LogicalKind identifies the normalized meaning of a database type.
type LogicalKind string

const (
	// KindUnknown preserves unrecognized native types without losing metadata.
	KindUnknown LogicalKind = "unknown"

	// KindString represents textual values.
	KindString LogicalKind = "string"

	// KindText represents large textual values.
	KindText LogicalKind = "text"

	// KindInteger represents signed or unsigned integer values.
	KindInteger LogicalKind = "integer"

	// KindDecimal represents fixed precision decimal values.
	KindDecimal LogicalKind = "decimal"

	// KindFloat represents floating point numeric values.
	KindFloat LogicalKind = "float"

	// KindBoolean represents true or false values.
	KindBoolean LogicalKind = "boolean"

	// KindDate represents calendar date values.
	KindDate LogicalKind = "date"

	// KindTime represents time-of-day values.
	KindTime LogicalKind = "time"

	// KindDateTime represents date and time values.
	KindDateTime LogicalKind = "datetime"

	// KindBinary represents binary byte values.
	KindBinary LogicalKind = "binary"

	// KindJSON represents structured JSON values.
	KindJSON LogicalKind = "json"

	// KindUUID represents UUID values.
	KindUUID LogicalKind = "uuid"

	// KindArray represents array values with an element type.
	KindArray LogicalKind = "array"

	// KindEnum represents constrained string-like enum values.
	KindEnum LogicalKind = "enum"
)

// LogicalType preserves normalized type semantics and the native type source.
type LogicalType struct {
	// Kind identifies the normalized type category.
	Kind LogicalKind

	// Length stores character, binary, or other length constraints when known.
	Length *int64

	// Precision stores numeric or temporal precision when known.
	Precision *int

	// Scale stores decimal scale when known.
	Scale *int

	// BitWidth stores numeric bit width when known.
	BitWidth *int

	// Timezone reports whether the type includes timezone semantics.
	Timezone bool

	// Element stores the element type for arrays and similar containers.
	Element *LogicalType

	// EnumValues stores the allowed values for enum-like types.
	EnumValues []string

	// NativeType stores the original database type name or definition.
	NativeType string
}

// UnknownLogicalType returns an unknown logical type while preserving the native type text.
func UnknownLogicalType(nativeType string) LogicalType {
	return LogicalType{Kind: KindUnknown, NativeType: nativeType}
}
