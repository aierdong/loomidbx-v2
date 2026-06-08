package storage

import "strings"

// DataClass identifies the local storage responsibility for a data category.
type DataClass string

const (
	// DataClassOrdinaryConfig is lightweight app configuration owned by the config module.
	DataClassOrdinaryConfig DataClass = "ordinary_config"

	// DataClassStructuredBusiness is local structured product data stored in SQLite by downstream specs.
	DataClassStructuredBusiness DataClass = "structured_business_data"

	// DataClassSensitiveSecret is credential material that must not be stored in config or business tables as plaintext.
	DataClassSensitiveSecret DataClass = "sensitive_secret"
)

// DataKind identifies product data categories documented by the storage policy.
type DataKind string

const (
	// DataTheme is a lightweight app appearance setting.
	DataTheme DataKind = "theme"

	// DataLanguage is a lightweight app language setting.
	DataLanguage DataKind = "language"

	// DataDataDir is the configured local data directory setting.
	DataDataDir DataKind = "data_dir"

	// DataDevelopmentOptions are development and test mode toggles.
	DataDevelopmentOptions DataKind = "development_options"

	// DataConnectionMetadata is non-secret database connection metadata.
	DataConnectionMetadata DataKind = "connection_metadata"

	// DataSchemaCache is locally cached database schema structure.
	DataSchemaCache DataKind = "schema_cache"

	// DataFieldRules are field generation rules configured by users.
	DataFieldRules DataKind = "field_rules"

	// DataProjectConfig is local project generation configuration.
	DataProjectConfig DataKind = "project_config"

	// DataExecutionHistory is local execution result and history data.
	DataExecutionHistory DataKind = "execution_history"

	// DataDatabasePassword is database credential plaintext.
	DataDatabasePassword DataKind = "database_password"

	// DataToken is token or API credential plaintext.
	DataToken DataKind = "token"
)

// StoragePolicy records local data classification and privacy boundary decisions.
type StoragePolicy struct {
	// Classification maps a data kind to its storage responsibility.
	Classification map[DataKind]DataClass

	// NetworkUploadDisabled confirms local product data is not uploaded by this storage layer.
	NetworkUploadDisabled bool
}

// DefaultPolicy returns the local storage classification required by this spec.
func DefaultPolicy() StoragePolicy {
	return StoragePolicy{
		Classification: map[DataKind]DataClass{
			DataTheme:              DataClassOrdinaryConfig,
			DataLanguage:           DataClassOrdinaryConfig,
			DataDataDir:            DataClassOrdinaryConfig,
			DataDevelopmentOptions: DataClassOrdinaryConfig,
			DataConnectionMetadata: DataClassStructuredBusiness,
			DataSchemaCache:        DataClassStructuredBusiness,
			DataFieldRules:         DataClassStructuredBusiness,
			DataProjectConfig:      DataClassStructuredBusiness,
			DataExecutionHistory:   DataClassStructuredBusiness,
			DataDatabasePassword:   DataClassSensitiveSecret,
			DataToken:              DataClassSensitiveSecret,
		},
		NetworkUploadDisabled: true,
	}
}

// Classify returns the storage responsibility for a data kind.
func (policy StoragePolicy) Classify(kind DataKind) DataClass {
	return policy.Classification[kind]
}

// RedactedText is the stable marker used when sensitive text is removed from errors or diagnostics.
const RedactedText = "[redacted]"

// RedactSensitive returns a stable redacted marker instead of potentially sensitive free text.
func RedactSensitive(value string) string {
	if value == "" {
		return RedactedText
	}
	lower := strings.ToLower(value)
	if strings.Contains(lower, "password") || strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "select ") || strings.Contains(lower, "user sql") {
		return RedactedText
	}
	return RedactedText
}
