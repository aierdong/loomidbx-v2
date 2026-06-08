package fakes

import (
	"strings"

	"github.com/gerdong/loomidbx/internal/dbx/schema"
	"github.com/gerdong/loomidbx/internal/dbx/typex"
)

// Mapper is a deterministic test-only native type mapper.
type Mapper struct {
	// Mappings stores configured logical types by lower-case native type text.
	Mappings map[string]schema.LogicalType

	// Calls stores the number of ToLogical invocations.
	Calls int

	// LastNative stores the last native type input.
	LastNative typex.NativeType

	// LastOptions stores the last mapping options input.
	LastOptions typex.MappingOptions
}

// NewMapper returns a fake mapper with a small deterministic mapping set.
func NewMapper() *Mapper {
	return &Mapper{Mappings: map[string]schema.LogicalType{
		"int":          {Kind: schema.KindInteger, NativeType: "int"},
		"integer":      {Kind: schema.KindInteger, NativeType: "integer"},
		"varchar":      {Kind: schema.KindString, NativeType: "varchar"},
		"varchar(255)": {Kind: schema.KindString, NativeType: "varchar(255)"},
		"text":         {Kind: schema.KindText, NativeType: "text"},
		"json":         {Kind: schema.KindJSON, NativeType: "json"},
	}}
}

// ToLogical maps native type text to configured logical types or an unknown fallback.
func (m *Mapper) ToLogical(native typex.NativeType, opts typex.MappingOptions) schema.LogicalType {
	m.Calls++
	m.LastNative = native
	m.LastOptions = opts
	nativeName := typex.NativeTypeName(native)
	if mapped, ok := m.Mappings[strings.ToLower(nativeName)]; ok {
		mapped.Length = native.Length
		mapped.Precision = native.Precision
		mapped.Scale = native.Scale
		if opts.PreserveNativeType || mapped.NativeType == "" {
			mapped.NativeType = nativeName
		}
		return mapped
	}
	kind := opts.Fallback
	if kind == "" {
		kind = schema.KindUnknown
	}
	return schema.LogicalType{Kind: kind, NativeType: nativeName, Length: native.Length, Precision: native.Precision, Scale: native.Scale}
}
