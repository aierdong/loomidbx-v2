package connection

// ParamValue stores a JSON-compatible extension parameter value.
type ParamValue any

// ConnectionParams stores extension parameters without interpreting them as driver behavior.
type ConnectionParams map[string]ParamValue
