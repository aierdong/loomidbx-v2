package storage

import "context"

// SecretRef is a persistable reference to credential material managed outside ordinary config and business tables.
type SecretRef struct {
	// Provider identifies the secret storage provider or boundary.
	Provider string

	// Key identifies the secret entry without containing the secret value.
	Key string
}

// Persistable reports whether this type may be stored in ordinary metadata records.
func (ref SecretRef) Persistable() bool {
	return ref.Provider != "" && ref.Key != ""
}

// SecretValue contains plaintext credential material only at the secret store call boundary.
type SecretValue struct {
	// Plaintext is credential material and must not be persisted in ordinary storage.
	Plaintext string
}

// Persistable reports whether this type may be stored in ordinary metadata records.
func (value SecretValue) Persistable() bool {
	return false
}

// SecretStore defines the future secure credential storage boundary.
type SecretStore interface {
	// Available reports whether a real secure secret store is configured.
	Available(ctx context.Context) bool

	// Get loads a secret value by reference or returns a typed unavailable error.
	Get(ctx context.Context, ref SecretRef) (SecretValue, error)

	// Put stores a secret value by reference or returns a typed unavailable error.
	Put(ctx context.Context, ref SecretRef, value SecretValue) error

	// Delete removes a secret value by reference or returns a typed unavailable error.
	Delete(ctx context.Context, ref SecretRef) error
}

// UnavailableSecretStore is the default implementation before platform secure storage exists.
type UnavailableSecretStore struct{}

// Available reports false because this implementation never stores plaintext credentials.
func (store UnavailableSecretStore) Available(ctx context.Context) bool {
	return false
}

// Get returns a stable unavailable error without leaking the referenced secret key.
func (store UnavailableSecretStore) Get(ctx context.Context, ref SecretRef) (SecretValue, error) {
	return SecretValue{}, secretUnavailableError("get secret")
}

// Put returns a stable unavailable error and never falls back to plaintext storage.
func (store UnavailableSecretStore) Put(ctx context.Context, ref SecretRef, value SecretValue) error {
	return secretUnavailableError("put secret")
}

// Delete returns a stable unavailable error without leaking the referenced secret key.
func (store UnavailableSecretStore) Delete(ctx context.Context, ref SecretRef) error {
	return secretUnavailableError("delete secret")
}

func secretUnavailableError(operation string) error {
	return NewError(ErrorCodeSecretStoreUnavailable, "secretRef", operation, "secure secret store is not configured", nil)
}
