package schema

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestRelationTypeStableStringValuesAndRecognition(t *testing.T) {
	tests := []struct {
		name         string
		relationType RelationType
		expected     string
	}{
		{name: "parent child", relationType: RelationTypeParentChild, expected: "PARENT_CHILD"},
		{name: "join table", relationType: RelationTypeJoinTable, expected: "JOIN_TABLE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.relationType.IsKnown() {
				t.Fatalf("%s should be recognized as a known relation type", tt.relationType)
			}
			if tt.relationType.IsUnknown() {
				t.Fatalf("%s should not be recognized as unknown", tt.relationType)
			}
			if got := tt.relationType.String(); got != tt.expected {
				t.Fatalf("String() = %q, want %q", got, tt.expected)
			}
		})
	}

	unknown := RelationType("FUTURE_RELATION")
	if unknown.IsKnown() {
		t.Fatalf("unknown relation type %q should not be known", unknown)
	}
	if !unknown.IsUnknown() {
		t.Fatalf("unknown relation type %q should be explicitly unknown", unknown)
	}
	if got := unknown.String(); got != "FUTURE_RELATION" {
		t.Fatalf("unknown relation type String() = %q, want unchanged value", got)
	}
}

func TestRelationTypeJSONRoundTripPreservesKnownAndUnknownValues(t *testing.T) {
	tests := []struct {
		name          string
		value         RelationType
		jsonValue     string
		expectedKnown bool
	}{
		{name: "parent child", value: RelationTypeParentChild, jsonValue: `"PARENT_CHILD"`, expectedKnown: true},
		{name: "join table", value: RelationTypeJoinTable, jsonValue: `"JOIN_TABLE"`, expectedKnown: true},
		{name: "unknown", value: RelationType("FUTURE_RELATION"), jsonValue: `"FUTURE_RELATION"`, expectedKnown: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Marshal(%v) returned error: %v", tt.value, err)
			}
			if string(encoded) != tt.jsonValue {
				t.Fatalf("Marshal(%v) = %s, want %s", tt.value, encoded, tt.jsonValue)
			}

			var decoded RelationType
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("Unmarshal(%s) returned error: %v", encoded, err)
			}
			if decoded.String() != strings.Trim(tt.jsonValue, `\"`) {
				t.Fatalf("decoded value = %q, want %s", decoded.String(), tt.jsonValue)
			}
			if decoded.IsKnown() != tt.expectedKnown {
				t.Fatalf("decoded known = %v, want %v", decoded.IsKnown(), tt.expectedKnown)
			}
		})
	}
}

func TestRelationTypeJSONRejectsNonStringValues(t *testing.T) {
	for _, raw := range []string{`1`, `true`, `{"relationType":"PARENT_CHILD"}`} {
		t.Run(raw, func(t *testing.T) {
			var decoded RelationType
			if err := json.Unmarshal([]byte(raw), &decoded); err == nil {
				t.Fatalf("Unmarshal RelationType should reject non-string JSON %s", raw)
			}
		})
	}
}

func TestValidateRelationTypeReturnsFieldLevelIssueForUnknownValues(t *testing.T) {
	if issues := ValidateRelationType(RelationTypeJoinTable); len(issues) != 0 {
		t.Fatalf("ValidateRelationType(known) = %#v, want no issues", issues)
	}

	issues := ValidateRelationType(RelationType("FUTURE_RELATION"))
	assertIssuePaths(t, issues, []string{"relationType"})
	assertIssueCodes(t, issues, map[string]SchemaIssueCode{"relationType": SchemaIssueCodeInvalidType})
	assertAllIssuesSafeErrors(t, issues)
}

func TestRelationMultiplicityValueObjectValidatesRange(t *testing.T) {
	multiplicity, issues := NewRelationMultiplicity(0, 3)
	if len(issues) != 0 {
		t.Fatalf("NewRelationMultiplicity(valid) issues = %#v, want none", issues)
	}
	if !reflect.DeepEqual(multiplicity, RelationMultiplicity{Min: 0, Max: 3}) {
		t.Fatalf("NewRelationMultiplicity(valid) = %#v, want Min=0 Max=3", multiplicity)
	}

	fixedZero, issues := NewRelationMultiplicity(0, 0)
	if len(issues) != 0 {
		t.Fatalf("NewRelationMultiplicity(0, 0) issues = %#v, want none", issues)
	}
	if fixedZero.Min != 0 || fixedZero.Max != 0 {
		t.Fatalf("NewRelationMultiplicity(0, 0) = %#v, want fixed zero range", fixedZero)
	}

	minIssues := ValidateRelationMultiplicity(-1, 0)
	assertIssuePaths(t, minIssues, []string{"multiplierMin"})
	assertIssueCodes(t, minIssues, map[string]SchemaIssueCode{"multiplierMin": SchemaIssueCodeInvalidRange})
	assertAllIssuesSafeErrors(t, minIssues)

	maxIssues := ValidateRelationMultiplicity(2, 1)
	assertIssuePaths(t, maxIssues, []string{"multiplierMax"})
	assertIssueCodes(t, maxIssues, map[string]SchemaIssueCode{"multiplierMax": SchemaIssueCodeInvalidRange})
	assertAllIssuesSafeErrors(t, maxIssues)
}

func TestRelationMultiplicityJSONSerializationIsSafeAndStable(t *testing.T) {
	original := RelationMultiplicity{Min: 1, Max: 5}
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal(RelationMultiplicity) returned error: %v", err)
	}
	if string(encoded) != `{"min":1,"max":5}` {
		t.Fatalf("RelationMultiplicity JSON = %s, want stable min/max fields", encoded)
	}

	var decoded RelationMultiplicity
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(RelationMultiplicity) returned error: %v", err)
	}
	if decoded != original {
		t.Fatalf("decoded RelationMultiplicity = %#v, want %#v", decoded, original)
	}
}
