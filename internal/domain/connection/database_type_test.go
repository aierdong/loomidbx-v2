package connection

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestDatabaseTypeKnownStableStringValues(t *testing.T) {
	tests := []struct {
		name     string
		dbType   DatabaseType
		expected string
	}{
		{name: "mysql", dbType: DatabaseTypeMySQL, expected: "mysql"},
		{name: "postgres", dbType: DatabaseTypePostgreSQL, expected: "postgres"},
		{name: "sqlite", dbType: DatabaseTypeSQLite, expected: "sqlite"},
		{name: "oracle", dbType: DatabaseTypeOracle, expected: "oracle"},
		{name: "sqlserver", dbType: DatabaseTypeSQLServer, expected: "sqlserver"},
		{name: "clickhouse", dbType: DatabaseTypeClickHouse, expected: "clickhouse"},
		{name: "tidb", dbType: DatabaseTypeTiDB, expected: "tidb"},
		{name: "hive", dbType: DatabaseTypeHive, expected: "hive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.dbType.IsKnown() {
				t.Fatalf("%s should be recognized as a known database type", tt.dbType)
			}
			if got := tt.dbType.String(); got != tt.expected {
				t.Fatalf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDatabaseTypeJSONRoundTripUsesStableStrings(t *testing.T) {
	tests := []struct {
		dbType DatabaseType
		json   string
	}{
		{dbType: DatabaseTypeMySQL, json: `"mysql"`},
		{dbType: DatabaseTypePostgreSQL, json: `"postgres"`},
		{dbType: DatabaseTypeSQLite, json: `"sqlite"`},
		{dbType: DatabaseTypeOracle, json: `"oracle"`},
		{dbType: DatabaseTypeSQLServer, json: `"sqlserver"`},
		{dbType: DatabaseTypeClickHouse, json: `"clickhouse"`},
		{dbType: DatabaseTypeTiDB, json: `"tidb"`},
		{dbType: DatabaseTypeHive, json: `"hive"`},
	}

	for _, tt := range tests {
		t.Run(tt.dbType.String(), func(t *testing.T) {
			encoded, err := json.Marshal(tt.dbType)
			if err != nil {
				t.Fatalf("Marshal(%s) returned error: %v", tt.dbType, err)
			}
			if string(encoded) != tt.json {
				t.Fatalf("Marshal(%s) = %s, want %s", tt.dbType, encoded, tt.json)
			}

			var decoded DatabaseType
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("Unmarshal(%s) returned error: %v", encoded, err)
			}
			if decoded != tt.dbType {
				t.Fatalf("round trip decoded %q, want %q", decoded, tt.dbType)
			}
			if !decoded.IsKnown() {
				t.Fatalf("decoded %q should remain known", decoded)
			}
		})
	}
}

func TestDatabaseTypeUnknownIsExplicitlyRecognized(t *testing.T) {
	unknown := DatabaseType("mariadb")

	if unknown.IsKnown() {
		t.Fatalf("unknown database type %q should not be known", unknown)
	}
	if got := unknown.String(); got != "mariadb" {
		t.Fatalf("String() = %q, want original unknown value", got)
	}

	encoded, err := json.Marshal(unknown)
	if err != nil {
		t.Fatalf("Marshal(%q) returned error: %v", unknown, err)
	}
	if string(encoded) != `"mariadb"` {
		t.Fatalf("Marshal(%q) = %s, want stable preservation", unknown, encoded)
	}

	var decoded DatabaseType
	if err := json.Unmarshal([]byte(`"mariadb"`), &decoded); err != nil {
		t.Fatalf("Unmarshal unknown database type returned error: %v", err)
	}
	if decoded.IsKnown() {
		t.Fatalf("decoded unknown database type %q should not be known", decoded)
	}
}

func TestDatabaseTypeBelongsToConnectionDomain(t *testing.T) {
	got := reflect.TypeOf(DatabaseTypeMySQL).PkgPath()
	want := "github.com/gerdong/loomidbx/internal/domain/connection"
	if got != want {
		t.Fatalf("DatabaseType package path = %q, want %q to keep it distinct from adapter DBType", got, want)
	}
}
