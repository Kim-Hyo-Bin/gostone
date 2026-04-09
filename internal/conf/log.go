package conf

// Log is the [log] INI section (HTTP access logging shape).
type Log struct {
	// JSON emits one JSON object per line instead of Gin’s colored text format.
	JSON bool `ini:"json"`
}
