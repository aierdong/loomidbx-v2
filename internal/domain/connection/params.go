package connection

import (
	"sort"
	"strings"
)

// ParamValue stores a JSON-compatible extension parameter value.
type ParamValue any

// ConnectionParams stores extension parameters without interpreting them as driver behavior.
type ConnectionParams map[string]ParamValue

// IsSensitiveParamKey reports whether a parameter key belongs to the sensitive credential boundary.
func IsSensitiveParamKey(key string) bool {
	lowerKey := strings.ToLower(key)
	for _, fragment := range []string{"password", "token", "secret", "credential"} {
		if strings.Contains(lowerKey, fragment) {
			return true
		}
	}
	return false
}

// SensitiveKeys returns extension parameter keys that must be handled through the sensitive credential boundary.
func (p ConnectionParams) SensitiveKeys() []string {
	keys := make([]string, 0)
	for key := range p {
		if IsSensitiveParamKey(key) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}
