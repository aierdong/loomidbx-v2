package connection

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestConnectionJSONRoundTripPreservesAggregateFields(t *testing.T) {
	original := Connection{
		ID:       ConnectionID("conn-reporting-primary"),
		Name:     "Reporting Primary",
		Type:     DatabaseTypeMySQL,
		Host:     "db.internal",
		Port:     3306,
		Database: "analytics",
		Username: "reporter",
		Credential: CredentialRef{
			ID:       CredentialID("cred-reporting-primary"),
			Type:     CredentialTypePassword,
			Provider: "local-secret-store",
			Key:      "connections/reporting-primary",
			Metadata: CredentialMetadata{"rotation": "manual"},
		},
		Params: ConnectionParams{
			"ssl":     true,
			"timeout": float64(30),
			"labels":  map[string]any{"env": "test"},
		},
	}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal(Connection) returned error: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("Unmarshal encoded connection into field map returned error: %v", err)
	}
	for _, field := range []string{"id", "name", "type", "host", "port", "database", "username", "credential", "params"} {
		if _, ok := fields[field]; !ok {
			t.Fatalf("encoded connection missing stable JSON field %q: %s", field, encoded)
		}
	}

	var decoded Connection
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(Connection) returned error: %v", err)
	}

	if decoded.ID != original.ID {
		t.Fatalf("decoded ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Name != original.Name {
		t.Fatalf("decoded Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Type != original.Type {
		t.Fatalf("decoded Type = %q, want %q", decoded.Type, original.Type)
	}
	if decoded.Host != original.Host {
		t.Fatalf("decoded Host = %q, want %q", decoded.Host, original.Host)
	}
	if decoded.Port != original.Port {
		t.Fatalf("decoded Port = %d, want %d", decoded.Port, original.Port)
	}
	if decoded.Database != original.Database {
		t.Fatalf("decoded Database = %q, want %q", decoded.Database, original.Database)
	}
	if decoded.Username != original.Username {
		t.Fatalf("decoded Username = %q, want %q", decoded.Username, original.Username)
	}
	if decoded.Credential.Provider != original.Credential.Provider {
		t.Fatalf("decoded credential provider = %q, want %q", decoded.Credential.Provider, original.Credential.Provider)
	}
	if decoded.Credential.Key != original.Credential.Key {
		t.Fatalf("decoded credential key = %q, want %q", decoded.Credential.Key, original.Credential.Key)
	}
	if !reflect.DeepEqual(decoded.Credential.Metadata, original.Credential.Metadata) {
		t.Fatalf("decoded credential metadata = %#v, want %#v", decoded.Credential.Metadata, original.Credential.Metadata)
	}
	if !reflect.DeepEqual(decoded.Params, original.Params) {
		t.Fatalf("decoded params = %#v, want %#v", decoded.Params, original.Params)
	}
}

func TestConnectionIDRemainsStableIdentityWhenNameChanges(t *testing.T) {
	first := Connection{ID: ConnectionID("conn-1"), Name: "Shared Name", Type: DatabaseTypePostgreSQL}
	second := Connection{ID: ConnectionID("conn-2"), Name: first.Name, Type: DatabaseTypePostgreSQL}

	if first.Name != second.Name {
		t.Fatalf("test setup expected matching names")
	}
	if first.ID == second.ID {
		t.Fatalf("connections with the same user-facing name should still have distinct IDs")
	}

	encoded, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("Marshal(Connection) returned error: %v", err)
	}
	if !strings.Contains(string(encoded), `"id":"conn-1"`) {
		t.Fatalf("encoded connection should preserve ID identity field: %s", encoded)
	}
}

func TestConnectionMissingOptionalFieldsUseSafeZeroValues(t *testing.T) {
	var decoded Connection
	if err := json.Unmarshal([]byte(`{"id":"conn-minimal","name":"Minimal","type":"sqlite"}`), &decoded); err != nil {
		t.Fatalf("Unmarshal minimal Connection returned error: %v", err)
	}

	if decoded.ID != ConnectionID("conn-minimal") {
		t.Fatalf("decoded ID = %q, want conn-minimal", decoded.ID)
	}
	if decoded.Name != "Minimal" {
		t.Fatalf("decoded Name = %q, want Minimal", decoded.Name)
	}
	if decoded.Type != DatabaseTypeSQLite {
		t.Fatalf("decoded Type = %q, want %q", decoded.Type, DatabaseTypeSQLite)
	}
	if decoded.Host != "" {
		t.Fatalf("missing host should decode to empty string, got %q", decoded.Host)
	}
	if decoded.Port != 0 {
		t.Fatalf("missing port should decode to zero, got %d", decoded.Port)
	}
	if decoded.Database != "" {
		t.Fatalf("missing database should decode to empty string, got %q", decoded.Database)
	}
	if decoded.Username != "" {
		t.Fatalf("missing username should decode to empty string, got %q", decoded.Username)
	}
	if !reflect.DeepEqual(decoded.Credential, CredentialRef{}) {
		t.Fatalf("missing credential should decode to zero reference, got %#v", decoded.Credential)
	}
	if decoded.Params != nil {
		t.Fatalf("missing params should decode to nil map, got %#v", decoded.Params)
	}
}

func TestConnectionNormalizeTrimsUserProvidedFields(t *testing.T) {
	conn := Connection{
		ID:       ConnectionID("conn-normalize"),
		Name:     "  Reporting Primary  ",
		Type:     DatabaseTypePostgreSQL,
		Host:     "  db.internal  ",
		Database: "  analytics  ",
		Username: "  reporter  ",
		Params: ConnectionParams{
			" sslmode ": " require ",
			"timeout":   float64(30),
		},
	}

	conn.Normalize()

	if conn.Name != "Reporting Primary" {
		t.Fatalf("normalized Name = %q, want Reporting Primary", conn.Name)
	}
	if conn.Host != "db.internal" {
		t.Fatalf("normalized Host = %q, want db.internal", conn.Host)
	}
	if conn.Database != "analytics" {
		t.Fatalf("normalized Database = %q, want analytics", conn.Database)
	}
	if conn.Username != "reporter" {
		t.Fatalf("normalized Username = %q, want reporter", conn.Username)
	}
	if _, ok := conn.Params[" sslmode "]; ok {
		t.Fatalf("normalization should trim parameter keys: %#v", conn.Params)
	}
	if got := conn.Params["sslmode"]; got != " require " {
		t.Fatalf("normalized param value = %#v, want original value", got)
	}
	if got := conn.Params["timeout"]; got != float64(30) {
		t.Fatalf("normalized timeout = %#v, want 30", got)
	}
}

func TestConnectionDoesNotExposePlaintextSecretFields(t *testing.T) {
	connectionType := reflect.TypeOf(Connection{})
	for i := range connectionType.NumField() {
		field := connectionType.Field(i)
		lowerName := strings.ToLower(field.Name)
		jsonName := strings.ToLower(strings.Split(field.Tag.Get("json"), ",")[0])
		for _, forbidden := range []string{"password", "token", "secret"} {
			if strings.Contains(lowerName, forbidden) || strings.Contains(jsonName, forbidden) {
				t.Fatalf("Connection exposes plaintext-like secret field %s with json tag %q", field.Name, field.Tag.Get("json"))
			}
		}
	}
}

func TestConnectionValidateAcceptsMinimalSQLiteConnection(t *testing.T) {
	conn := Connection{Name: "Local", Type: DatabaseTypeSQLite}

	result := conn.Validate()

	if len(result.Errors) != 0 {
		t.Fatalf("Validate() errors = %#v, want none", result.Errors)
	}
}

func TestConnectionValidateReportsAllBasicErrors(t *testing.T) {
	conn := Connection{
		Name: "  ",
		Type: DatabaseType("mariadb"),
		Port: 70000,
		Params: ConnectionParams{
			" ":         "blank key",
			"password":  "do-not-log",
			"api_token": "do-not-log-either",
		},
	}

	result := conn.Validate()

	expected := []ValidationError{
		{Field: "name", Code: ValidationCodeRequired, Message: "connection name is required"},
		{Field: "type", Code: ValidationCodeUnknownDatabaseType, Message: "database type is not recognized"},
		{Field: "port", Code: ValidationCodeInvalidPort, Message: "port must be between 1 and 65535"},
		{Field: "params", Code: ValidationCodeRequired, Message: "parameter key is required"},
		{Field: "params.api_token", Code: ValidationCodeSensitiveParamNotAllowed, Message: "sensitive parameter must use credential reference"},
		{Field: "params.password", Code: ValidationCodeSensitiveParamNotAllowed, Message: "sensitive parameter must use credential reference"},
	}
	if !reflect.DeepEqual(result.Errors, expected) {
		t.Fatalf("Validate() errors = %#v, want %#v", result.Errors, expected)
	}
	for _, validationError := range result.Errors {
		for _, leaked := range []string{"do-not-log", "do-not-log-either"} {
			if strings.Contains(validationError.Message, leaked) {
				t.Fatalf("validation error leaked sensitive value in message: %#v", validationError)
			}
		}
	}
}

func TestConnectionValidateRequiresHostForNetworkDatabases(t *testing.T) {
	conn := Connection{Name: "Reporting", Type: DatabaseTypePostgreSQL}

	result := conn.Validate()

	expected := []ValidationError{{Field: "host", Code: ValidationCodeRequired, Message: "host is required for network database types"}}
	if !reflect.DeepEqual(result.Errors, expected) {
		t.Fatalf("Validate() errors = %#v, want %#v", result.Errors, expected)
	}
}

func TestConnectionUnknownJSONFieldsDoNotBreakKnownFieldRecovery(t *testing.T) {
	payload := []byte(`{"id":"conn-compat","name":"Compatible","type":"mysql","host":"db.internal","future_field":{"nested":true}}`)

	var decoded Connection
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("Unmarshal(Connection) with unknown fields returned error: %v", err)
	}

	if decoded.ID != ConnectionID("conn-compat") {
		t.Fatalf("decoded ID = %q, want conn-compat", decoded.ID)
	}
	if decoded.Name != "Compatible" {
		t.Fatalf("decoded Name = %q, want Compatible", decoded.Name)
	}
	if decoded.Type != DatabaseTypeMySQL {
		t.Fatalf("decoded Type = %q, want %q", decoded.Type, DatabaseTypeMySQL)
	}
	if decoded.Host != "db.internal" {
		t.Fatalf("decoded Host = %q, want db.internal", decoded.Host)
	}
}

func TestConnectionValidationErrorsExcludeSensitiveParamValues(t *testing.T) {
	conn := Connection{
		Name: "Sensitive Params",
		Type: DatabaseTypeMySQL,
		Host: "db.internal",
		Params: ConnectionParams{
			"password":     "do-not-report-password",
			"access_token": "do-not-report-token",
		},
	}

	result := conn.Validate()

	if len(result.Errors) == 0 {
		t.Fatalf("Validate() should report sensitive parameter key errors")
	}
	for _, validationError := range result.Errors {
		for _, leaked := range []string{"do-not-report-password", "do-not-report-token"} {
			if strings.Contains(validationError.Message, leaked) {
				t.Fatalf("validation error leaked sensitive parameter value %q: %#v", leaked, validationError)
			}
		}
	}
}
