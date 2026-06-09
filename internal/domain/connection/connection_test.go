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
