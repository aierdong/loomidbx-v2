package connection

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestValidationErrorJSONUsesStableFieldCodeAndMessage(t *testing.T) {
	validationError := ValidationError{
		Field:   "name",
		Code:    ValidationCodeRequired,
		Message: "connection name is required",
	}

	encoded, err := json.Marshal(validationError)
	if err != nil {
		t.Fatalf("Marshal(ValidationError) returned error: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("Unmarshal encoded validation error into field map returned error: %v", err)
	}
	for _, field := range []string{"field", "code", "message"} {
		if _, ok := fields[field]; !ok {
			t.Fatalf("encoded validation error missing stable JSON field %q: %s", field, encoded)
		}
	}

	var decoded ValidationError
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(ValidationError) returned error: %v", err)
	}

	if decoded != validationError {
		t.Fatalf("decoded validation error = %#v, want %#v", decoded, validationError)
	}
}

func TestValidationResultPreservesMultipleFieldErrors(t *testing.T) {
	original := ValidationResult{
		Errors: []ValidationError{
			{Field: "name", Code: ValidationCodeRequired, Message: "connection name is required"},
			{Field: "type", Code: ValidationCodeUnknownDatabaseType, Message: "database type is not recognized"},
			{Field: "port", Code: ValidationCodeInvalidPort, Message: "port must be between 1 and 65535"},
			{Field: "params.password", Code: ValidationCodeSensitiveParamNotAllowed, Message: "sensitive parameter must use credential reference"},
		},
	}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal(ValidationResult) returned error: %v", err)
	}

	var decoded ValidationResult
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(ValidationResult) returned error: %v", err)
	}

	if !reflect.DeepEqual(decoded, original) {
		t.Fatalf("decoded validation result = %#v, want %#v", decoded, original)
	}
}

func TestValidationCodesUseStableStringValues(t *testing.T) {
	tests := []struct {
		code     ValidationCode
		expected string
	}{
		{code: ValidationCodeRequired, expected: "required"},
		{code: ValidationCodeUnknownDatabaseType, expected: "unknown_database_type"},
		{code: ValidationCodeInvalidPort, expected: "invalid_port"},
		{code: ValidationCodeSensitiveParamNotAllowed, expected: "sensitive_param_not_allowed"},
	}

	for _, tt := range tests {
		if string(tt.code) != tt.expected {
			t.Fatalf("ValidationCode = %q, want %q", tt.code, tt.expected)
		}
	}
}
