package connection

// CredentialID is the stable business identity for a credential reference.
type CredentialID string

// CredentialType describes the kind of credential material referenced by a connection.
type CredentialType string

const (
	// CredentialTypePassword identifies password credential material stored outside ordinary business data.
	CredentialTypePassword CredentialType = "password"

	// CredentialTypeToken identifies token credential material stored outside ordinary business data.
	CredentialTypeToken CredentialType = "token"
)

// CredentialMetadata stores non-secret metadata about a credential reference.
type CredentialMetadata map[string]string

// CredentialRef references sensitive credential material without containing plaintext secrets.
type CredentialRef struct {
	// ID stores the stable business identity of the credential reference.
	ID CredentialID `json:"id,omitempty"`

	// Type stores the kind of credential material referenced by this value.
	Type CredentialType `json:"type,omitempty"`

	// Provider identifies the secret storage provider or boundary for later storage mapping.
	Provider string `json:"provider,omitempty"`

	// Key identifies the secret storage entry without containing the secret value.
	Key string `json:"key,omitempty"`

	// Metadata stores non-secret credential metadata that does not belong in the secret store reference.
	Metadata CredentialMetadata `json:"metadata,omitempty"`
}
