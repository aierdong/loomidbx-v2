package connection

import (
	"sort"
	"strings"
)

// ConnectionID is the stable identity used to reference a saved connection.
type ConnectionID string

// Connection represents a reusable database connection configuration aggregate.
type Connection struct {
	// ID stores the stable connection identity for downstream references.
	ID ConnectionID `json:"id"`

	// Name stores the user-facing connection name and is not the unique identity.
	Name string `json:"name"`

	// Type stores the business database type selected for this connection.
	Type DatabaseType `json:"type"`

	// Host stores the database host for network database types.
	Host string `json:"host,omitempty"`

	// Port stores the database TCP port when a port is applicable.
	Port int `json:"port,omitempty"`

	// Database stores the initial database, catalog, or namespace for the connection.
	Database string `json:"database,omitempty"`

	// Username stores the non-secret login name for the connection.
	Username string `json:"username,omitempty"`

	// Credential stores the reference to secret credential material outside ordinary business data.
	Credential CredentialRef `json:"credential,omitempty"`

	// Params stores non-core extension parameters without interpreting them as driver behavior.
	Params ConnectionParams `json:"params,omitempty"`
}

// Normalize trims user-provided connection fields and parameter keys in place.
func (c *Connection) Normalize() {
	c.Name = strings.TrimSpace(c.Name)
	c.Host = strings.TrimSpace(c.Host)
	c.Database = strings.TrimSpace(c.Database)
	c.Username = strings.TrimSpace(c.Username)

	if c.Params == nil {
		return
	}

	normalized := make(ConnectionParams, len(c.Params))
	for key, value := range c.Params {
		normalized[strings.TrimSpace(key)] = value
	}
	c.Params = normalized
}

// Validate returns all detectable field-level validation errors for the connection.
func (c Connection) Validate() ValidationResult {
	result := ValidationResult{}

	if strings.TrimSpace(c.Name) == "" {
		result.Errors = append(result.Errors, ValidationError{Field: "name", Code: ValidationCodeRequired, Message: "connection name is required"})
	}
	if !c.Type.IsKnown() {
		result.Errors = append(result.Errors, ValidationError{Field: "type", Code: ValidationCodeUnknownDatabaseType, Message: "database type is not recognized"})
	}
	if c.Type.RequiresNetworkAddress() && strings.TrimSpace(c.Host) == "" {
		result.Errors = append(result.Errors, ValidationError{Field: "host", Code: ValidationCodeRequired, Message: "host is required for network database types"})
	}
	if c.Port < 0 || c.Port > 65535 {
		result.Errors = append(result.Errors, ValidationError{Field: "port", Code: ValidationCodeInvalidPort, Message: "port must be between 1 and 65535"})
	}

	keys := make([]string, 0, len(c.Params))
	for key := range c.Params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			result.Errors = append(result.Errors, ValidationError{Field: "params", Code: ValidationCodeRequired, Message: "parameter key is required"})
			continue
		}
		if IsSensitiveParamKey(key) {
			result.Errors = append(result.Errors, ValidationError{Field: "params." + trimmedKey, Code: ValidationCodeSensitiveParamNotAllowed, Message: "sensitive parameter must use credential reference"})
		}
	}

	return result
}
