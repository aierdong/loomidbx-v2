package storage_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gerdong/loomidbx/internal/storage"
)

func TestUnavailableSecretStoreDoesNotFallbackToPlaintext(t *testing.T) {
	ctx := context.Background()
	store := storage.UnavailableSecretStore{}
	ref := storage.SecretRef{Provider: "keychain", Key: "connection/main"}
	secret := storage.SecretValue{Plaintext: "password=secret-token"}

	if store.Available(ctx) {
		t.Fatal("Available() = true, want false")
	}
	if _, err := store.Get(ctx, ref); !isSecretUnavailable(err) {
		t.Fatalf("Get() error = %T %[1]v, want SECRET_STORE_UNAVAILABLE", err)
	}
	if err := store.Put(ctx, ref, secret); !isSecretUnavailable(err) {
		t.Fatalf("Put() error = %T %[1]v, want SECRET_STORE_UNAVAILABLE", err)
	}
	if err := store.Delete(ctx, ref); !isSecretUnavailable(err) {
		t.Fatalf("Delete() error = %T %[1]v, want SECRET_STORE_UNAVAILABLE", err)
	}
}

func TestSecretErrorsAndDiagnosticsAreRedacted(t *testing.T) {
	ctx := context.Background()
	store := storage.UnavailableSecretStore{}
	ref := storage.SecretRef{Provider: "keychain", Key: "connection/password-secret-token"}

	_, err := store.Get(ctx, ref)
	var storageErr *storage.StorageError
	if !errors.As(err, &storageErr) {
		t.Fatalf("Get() error = %T %[1]v, want StorageError", err)
	}
	text := storageErr.Error()
	for _, forbidden := range []string{"password-secret-token", "secret-token", "password="} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("Error() = %q, want forbidden sensitive value %q redacted", text, forbidden)
		}
	}
	if storageErr.Field != "secretRef" {
		t.Fatalf("Field = %q, want secretRef", storageErr.Field)
	}
}

func TestSecretRefIsPersistableButSecretValueIsNot(t *testing.T) {
	ref := storage.SecretRef{Provider: "keychain", Key: "connection/main"}
	value := storage.SecretValue{Plaintext: "super-secret"}

	if !ref.Persistable() {
		t.Fatal("SecretRef.Persistable() = false, want true")
	}
	if value.Persistable() {
		t.Fatal("SecretValue.Persistable() = true, want false")
	}
	if storage.RedactSensitive("token=abc user sql select * from t") != storage.RedactedText {
		t.Fatal("RedactSensitive() did not return stable redacted marker")
	}
}

func isSecretUnavailable(err error) bool {
	var storageErr *storage.StorageError
	return errors.As(err, &storageErr) && storageErr.Code == storage.ErrorCodeSecretStoreUnavailable
}
