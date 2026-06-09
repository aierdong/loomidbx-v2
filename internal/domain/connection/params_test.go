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
