package connection

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestConnectionParamsJSONRoundTripPreservesCompatibleValues(t *testing.T) {
	original := ConnectionParams{
		"ssl":     true,
		"timeout": float64(30),
		"charset": "utf8mb4",
		"labels": map[string]any{
			"env":      "test",
			"priority": float64(1),
		},
		"replicas": []any{"db-replica-1", "db-replica-2"},
	}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal(ConnectionParams) returned error: %v", err)
	}

	var decoded ConnectionParams
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(ConnectionParams) returned error: %v", err)
	}

	if !reflect.DeepEqual(decoded, original) {
		t.Fatalf("decoded params = %#v, want %#v", decoded, original)
	}
}

func TestConnectionParamsPreserveKeysWithoutDriverInterpretation(t *testing.T) {
	original := ConnectionParams{
		"sslmode":              "require",
		"mysql_parse_time":     true,
		"application_name":     "loomidbx",
		"connect_timeout_secs": float64(5),
	}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal(ConnectionParams) returned error: %v", err)
	}

	var decoded ConnectionParams
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(ConnectionParams) returned error: %v", err)
	}

	for key, want := range original {
		if got := decoded[key]; !reflect.DeepEqual(got, want) {
			t.Fatalf("decoded param %q = %#v, want %#v", key, got, want)
		}
	}
}

func TestConnectionParamKeyIdentifiesSensitiveFragments(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{key: "password", expected: true},
		{key: "dbPassword", expected: true},
		{key: "access_token", expected: true},
		{key: "clientSecret", expected: true},
		{key: "credential_ref", expected: true},
		{key: " SSLMode ", expected: false},
		{key: "application_name", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := IsSensitiveParamKey(tt.key); got != tt.expected {
				t.Fatalf("IsSensitiveParamKey(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestConnectionParamsSensitiveKeysReturnsOnlySensitiveKeys(t *testing.T) {
	params := ConnectionParams{
		"sslmode":        "require",
		"Password":       "do-not-log-this",
		"api_token":      "do-not-log-this-either",
		"connectTimeout": float64(10),
	}

	sensitiveKeys := params.SensitiveKeys()

	if !reflect.DeepEqual(sensitiveKeys, []string{"Password", "api_token"}) {
		t.Fatalf("SensitiveKeys() = %#v, want sensitive keys only", sensitiveKeys)
	}
}
