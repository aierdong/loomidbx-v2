package schema

import "encoding/json"

// ColumnLogicalKind identifies the stable logical type category for a database column.
type ColumnLogicalKind string

const (
	// ColumnLogicalKindUnknown represents an unrecognized database type that must preserve NativeType.
	ColumnLogicalKindUnknown ColumnLogicalKind = "unknown"

	// ColumnLogicalKindString represents bounded short text values.
	ColumnLogicalKindString ColumnLogicalKind = "string"

	// ColumnLogicalKindText represents long text values.
	ColumnLogicalKindText ColumnLogicalKind = "text"

	// ColumnLogicalKindInteger represents integral numeric values.
	ColumnLogicalKindInteger ColumnLogicalKind = "integer"

	// ColumnLogicalKindDecimal represents fixed-point numeric values.
	ColumnLogicalKindDecimal ColumnLogicalKind = "decimal"

	// ColumnLogicalKindFloat represents floating-point numeric values.
	ColumnLogicalKindFloat ColumnLogicalKind = "float"

	// ColumnLogicalKindBoolean represents boolean values.
	ColumnLogicalKindBoolean ColumnLogicalKind = "boolean"

	// ColumnLogicalKindDate represents calendar date values without time of day.
	ColumnLogicalKindDate ColumnLogicalKind = "date"

	// ColumnLogicalKindTime represents time-of-day values without calendar date.
	ColumnLogicalKindTime ColumnLogicalKind = "time"

	// ColumnLogicalKindDateTime represents combined date and time values.
	ColumnLogicalKindDateTime ColumnLogicalKind = "datetime"

	// ColumnLogicalKindBinary represents binary byte values.
	ColumnLogicalKindBinary ColumnLogicalKind = "binary"

	// ColumnLogicalKindJSON represents JSON document values.
	ColumnLogicalKindJSON ColumnLogicalKind = "json"

	// ColumnLogicalKindUUID represents UUID identifier values.
	ColumnLogicalKindUUID ColumnLogicalKind = "uuid"

	// ColumnLogicalKindArray represents array values whose element type may be described recursively.
	ColumnLogicalKindArray ColumnLogicalKind = "array"

	// ColumnLogicalKindEnum represents enumerated string values.
	ColumnLogicalKindEnum ColumnLogicalKind = "enum"
)

// IsKnown reports whether the logical kind belongs to the stable supported set.
func (kind ColumnLogicalKind) IsKnown() bool {
	switch kind {
	case ColumnLogicalKindUnknown,
		ColumnLogicalKindString,
		ColumnLogicalKindText,
		ColumnLogicalKindInteger,
		ColumnLogicalKindDecimal,
		ColumnLogicalKindFloat,
		ColumnLogicalKindBoolean,
		ColumnLogicalKindDate,
		ColumnLogicalKindTime,
		ColumnLogicalKindDateTime,
		ColumnLogicalKindBinary,
		ColumnLogicalKindJSON,
		ColumnLogicalKindUUID,
		ColumnLogicalKindArray,
		ColumnLogicalKindEnum:
		return true
	default:
		return false
	}
}

// IsUnknown reports whether the logical kind is outside the stable supported set.
func (kind ColumnLogicalKind) IsUnknown() bool {
	return !kind.IsKnown()
}

// String returns the stable string representation used for persistence and transport.
func (kind ColumnLogicalKind) String() string {
	return string(kind)
}

// MarshalJSON serializes the logical kind as its stable string value.
func (kind ColumnLogicalKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}

// UnmarshalJSON restores the logical kind from its serialized string value.
func (kind *ColumnLogicalKind) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*kind = ColumnLogicalKind(value)
	return nil
}

// ColumnLogicalType describes stable logical type metadata for downstream field rules and generators.
type ColumnLogicalType struct {
	// Kind stores the stable logical type category.
	Kind ColumnLogicalKind `json:"kind"`

	// Length stores an optional positive length for bounded character or binary types.
	Length *int64 `json:"length"`

	// Precision stores an optional positive precision for numeric or temporal types.
	Precision *int `json:"precision"`

	// Scale stores an optional non-negative decimal scale that must not exceed Precision when both are known.
	Scale *int `json:"scale"`

	// BitWidth stores an optional positive bit width for numeric values.
	BitWidth *int `json:"bitWidth"`

	// Timezone reports whether the logical temporal type carries timezone semantics.
	Timezone bool `json:"timezone"`

	// Element stores optional recursive element metadata for array logical types.
	Element *ColumnLogicalType `json:"element"`

	// EnumValues stores optional stable values for enum logical types.
	EnumValues []string `json:"enumValues"`

	// NativeType preserves the database-native type text, especially for unknown logical types.
	NativeType string `json:"nativeType"`
}
