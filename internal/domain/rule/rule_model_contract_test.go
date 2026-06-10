package rule

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGeneratorConfigStableIdentityParentReferencesFieldsAndJSONTags(t *testing.T) {
	configType := reflect.TypeOf(GeneratorConfig{})

	assertRuleJSONTags(t, configType, map[string]string{
		"ID":              "id",
		"ColumnID":        "columnId",
		"GeneratorName":   "generatorName",
		"DataMappingType": "dataMappingType",
		"Params":          "params",
		"ConfigStatus":    "configStatus",
		"CreatedAt":       "createdAt",
		"UpdatedAt":       "updatedAt",
	})
	assertRuleStructJSONFieldSet(t, configType, []string{"id", "columnId", "generatorName", "dataMappingType", "params", "configStatus", "createdAt", "updatedAt"})

	assertRuleFieldType(t, configType, "ID", reflect.TypeOf(int64(0)))
	assertRuleFieldType(t, configType, "ColumnID", reflect.TypeOf(int64(0)))
	assertRuleFieldType(t, configType, "GeneratorName", reflect.TypeOf(""))
	assertRuleFieldType(t, configType, "DataMappingType", reflect.TypeOf(DataMappingType("")))
	assertRuleFieldType(t, configType, "Params", reflect.TypeOf(GeneratorParams{}))
	assertRuleFieldType(t, configType, "ConfigStatus", reflect.TypeOf(ConfigStatus("")))
	assertRuleFieldType(t, configType, "CreatedAt", reflect.TypeOf(time.Time{}))
	assertRuleFieldType(t, configType, "UpdatedAt", reflect.TypeOf(time.Time{}))

	createdAt := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)
	config := GeneratorConfig{
		ID:              101,
		ColumnID:        202,
		GeneratorName:   "person.name",
		DataMappingType: DataMappingTypeText,
		Params:          GeneratorParams{},
		ConfigStatus:    ConfigStatusActive,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}

	encoded, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Marshal(GeneratorConfig) returned error: %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("Unmarshal encoded GeneratorConfig into field map returned error: %v", err)
	}
	assertRuleJSONFieldsPresent(t, fields, "id", "columnId", "generatorName", "dataMappingType", "params", "configStatus", "createdAt", "updatedAt")
	assertRuleJSONField(t, fields, "id", "101")
	assertRuleJSONField(t, fields, "columnId", "202")
	assertRuleJSONField(t, fields, "generatorName", `"person.name"`)
	assertRuleJSONField(t, fields, "dataMappingType", `"text"`)
	assertRuleJSONField(t, fields, "configStatus", `"ACTIVE"`)

	var decoded GeneratorConfig
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(GeneratorConfig) returned error: %v", err)
	}
	if decoded.ID != config.ID || decoded.ColumnID != config.ColumnID || decoded.GeneratorName != config.GeneratorName || decoded.DataMappingType != config.DataMappingType || decoded.ConfigStatus != config.ConfigStatus {
		t.Fatalf("decoded GeneratorConfig identity and core fields = %#v, want %#v", decoded, config)
	}
}

func TestDataMappingTypeStableStringValuesRecognitionAndJSON(t *testing.T) {
	tests := []struct {
		name      string
		value     DataMappingType
		jsonValue string
	}{
		{name: "text", value: DataMappingTypeText, jsonValue: `"text"`},
		{name: "integer", value: DataMappingTypeInteger, jsonValue: `"integer"`},
		{name: "float", value: DataMappingTypeFloat, jsonValue: `"float"`},
		{name: "boolean", value: DataMappingTypeBoolean, jsonValue: `"boolean"`},
		{name: "datetime", value: DataMappingTypeDatetime, jsonValue: `"datetime"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.value.IsKnown() {
				t.Fatalf("%q should be recognized as known", tt.value)
			}
			if tt.value.IsUnknown() {
				t.Fatalf("%q should not be recognized as unknown", tt.value)
			}
			if got := tt.value.String(); got != strings.Trim(tt.jsonValue, `"`) {
				t.Fatalf("String() = %q, want %s", got, tt.jsonValue)
			}

			encoded, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Marshal(%v) returned error: %v", tt.value, err)
			}
			if string(encoded) != tt.jsonValue {
				t.Fatalf("Marshal(%v) = %s, want %s", tt.value, encoded, tt.jsonValue)
			}

			var decoded DataMappingType
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("Unmarshal(%s) returned error: %v", encoded, err)
			}
			if decoded != tt.value {
				t.Fatalf("decoded DataMappingType = %q, want %q", decoded, tt.value)
			}
		})
	}

	unknown := DataMappingType("json")
	if unknown.IsKnown() {
		t.Fatalf("unknown data mapping type %q should not be known", unknown)
	}
	if !unknown.IsUnknown() {
		t.Fatalf("unknown data mapping type %q should be explicitly unknown", unknown)
	}
	if got := unknown.String(); got != "json" {
		t.Fatalf("unknown data mapping type String() = %q, want unchanged value", got)
	}
	encoded, err := json.Marshal(unknown)
	if err != nil {
		t.Fatalf("Marshal(unknown DataMappingType) returned error: %v", err)
	}
	if string(encoded) != `"json"` {
		t.Fatalf("unknown DataMappingType JSON = %s, want preserved string", encoded)
	}
}

func TestGeneratorConfigExcludesOutOfScopeFields(t *testing.T) {
	configType := reflect.TypeOf(GeneratorConfig{})
	for index := range configType.NumField() {
		field := configType.Field(index)
		fieldName := strings.ToLower(field.Name)
		jsonName := strings.ToLower(strings.Split(field.Tag.Get("json"), ",")[0])
		for _, forbidden := range []string{"service", "api", "ui", "wails", "vue", "execution", "engine", "driver", "sql", "project", "row", "clear", "preview", "registry", "database", "relation", "order", "sort", "algorithm"} {
			if strings.Contains(fieldName, forbidden) || strings.Contains(jsonName, forbidden) {
				t.Fatalf("GeneratorConfig.%s exposes out-of-scope field matching %q with json tag %q", field.Name, forbidden, field.Tag.Get("json"))
			}
		}
	}
}

func assertRuleJSONTags(t *testing.T, typ reflect.Type, expected map[string]string) {
	t.Helper()

	for fieldName, expectedTag := range expected {
		field, ok := typ.FieldByName(fieldName)
		if !ok {
			t.Fatalf("%s missing field %s", typ.Name(), fieldName)
		}
		if got := strings.Split(field.Tag.Get("json"), ",")[0]; got != expectedTag {
			t.Fatalf("%s.%s json tag = %q, want %q", typ.Name(), fieldName, got, expectedTag)
		}
	}
}

func assertRuleStructJSONFieldSet(t *testing.T, typ reflect.Type, expected []string) {
	t.Helper()

	actual := make([]string, 0, typ.NumField())
	for index := range typ.NumField() {
		field := typ.Field(index)
		jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonName == "" || jsonName == "-" {
			continue
		}
		actual = append(actual, jsonName)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%s JSON field set = %#v, want %#v", typ.Name(), actual, expected)
	}
}

func assertRuleFieldType(t *testing.T, typ reflect.Type, fieldName string, expected reflect.Type) {
	t.Helper()

	field, ok := typ.FieldByName(fieldName)
	if !ok {
		t.Fatalf("%s missing field %s", typ.Name(), fieldName)
	}
	if field.Type != expected {
		t.Fatalf("%s.%s type = %s, want %s", typ.Name(), fieldName, field.Type, expected)
	}
}

func assertRuleJSONFieldsPresent(t *testing.T, fields map[string]json.RawMessage, expected ...string) {
	t.Helper()

	if len(fields) != len(expected) {
		t.Fatalf("JSON field count = %d (%v), want %d (%v)", len(fields), fields, len(expected), expected)
	}
	for _, field := range expected {
		if _, ok := fields[field]; !ok {
			t.Fatalf("encoded JSON missing field %q in %v", field, fields)
		}
	}
}

func assertRuleJSONField(t *testing.T, fields map[string]json.RawMessage, field string, expected string) {
	t.Helper()

	raw, ok := fields[field]
	if !ok {
		t.Fatalf("encoded JSON missing field %q", field)
	}
	if got := string(raw); got != expected {
		t.Fatalf("field %q JSON = %s, want %s", field, got, expected)
	}
}
