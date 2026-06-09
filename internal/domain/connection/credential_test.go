package connection

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/gerdong/loomidbx/internal/storage"
)

func TestCredentialRefJSONRoundTripPreservesReferenceFields(t *testing.T) {
	original := CredentialRef{
		ID:       CredentialID("cred-reporting-primary"),
		Type:     CredentialTypePassword,
		Provider: "local-secret-store",
		Key:      "connections/reporting-primary",
		Metadata: CredentialMetadata{"rotation": "manual"},
	}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal(CredentialRef) returned error: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("Unmarshal encoded credential reference into field map returned error: %v", err)
	}
	for _, field := range []string{"id", "type", "provider", "key", "metadata"} {
		if _, ok := fields[field]; !ok {
			t.Fatalf("encoded credential reference missing stable JSON field %q: %s", field, encoded)
		}
	}

	var decoded CredentialRef
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal(CredentialRef) returned error: %v", err)
	}

	if !reflect.DeepEqual(decoded, original) {
		t.Fatalf("decoded credential reference = %#v, want %#v", decoded, original)
	}
}

func TestCredentialRefMapsLosslesslyToStorageSecretRef(t *testing.T) {
	credential := CredentialRef{
		ID:       CredentialID("cred-business-id"),
		Type:     CredentialTypeToken,
		Provider: "keychain",
		Key:      "connections/reporting/token",
		Metadata: CredentialMetadata{"scope": "read-only"},
	}

	secretRef := storage.SecretRef{Provider: credential.Provider, Key: credential.Key}

	if secretRef.Provider != credential.Provider {
		t.Fatalf("SecretRef.Provider = %q, want %q", secretRef.Provider, credential.Provider)
	}
	if secretRef.Key != credential.Key {
		t.Fatalf("SecretRef.Key = %q, want %q", secretRef.Key, credential.Key)
	}
	if !secretRef.Persistable() {
		t.Fatalf("SecretRef derived from credential provider/key should be persistable")
	}
}

func TestCredentialRefDoesNotExposePlaintextSecretFields(t *testing.T) {
	credentialType := reflect.TypeOf(CredentialRef{})
	for i := range credentialType.NumField() {
		field := credentialType.Field(i)
		lowerName := strings.ToLower(field.Name)
		jsonName := strings.ToLower(strings.Split(field.Tag.Get("json"), ",")[0])
		for _, forbidden := range []string{"plaintext", "value", "password", "token", "secret"} {
			if strings.Contains(lowerName, forbidden) || strings.Contains(jsonName, forbidden) {
				t.Fatalf("CredentialRef exposes plaintext-like secret field %s with json tag %q", field.Name, field.Tag.Get("json"))
			}
		}
	}
}
