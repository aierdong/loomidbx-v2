package typex

import "github.com/gerdong/loomidbx/internal/dbx/schema"

// NativeType describes a database-native type observed during metadata scanning.
type NativeType struct {
	// Name stores the database type base name.
	Name string

	// Full stores the complete native type expression.
	Full string

	// Length stores length metadata when present.
	Length *int64

	// Precision stores numeric or temporal precision when present.
	Precision *int

	// Scale stores decimal scale when present.
	Scale *int

	// Nullable reports whether the source column allows null values.
	Nullable bool

	// Raw stores unnormalized native type metadata.
	Raw map[string]any
}

// MappingOptions controls native-to-logical fallback and adapter-specific mapping behavior.
type MappingOptions struct {
	// Fallback stores the logical kind to use when the mapper cannot recognize a type.
	Fallback schema.LogicalKind

	// PreserveNativeType reports whether the returned logical type should keep native text.
	PreserveNativeType bool

	// Options stores adapter-specific mapping switches.
	Options map[string]any
}

// Mapper converts native database types to canonical logical types without opening connections.
type Mapper interface {
	// ToLogical maps a native database type into a logical type.
	ToLogical(native NativeType, opts MappingOptions) schema.LogicalType
}

// NativeTypeName returns the best available source type text for preservation.
func NativeTypeName(native NativeType) string {
	if native.Full != "" {
		return native.Full
	}
	return native.Name
}
