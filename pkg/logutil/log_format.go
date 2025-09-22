package logutil

import "fmt"

const (
	JSONLogFormat    = "json"
	ConsoleLogFormat = "console"
	//revive:disable:var-naming
	// Deprecated: Please use JSONLogFormat.
	JsonLogFormat = JSONLogFormat
	//revive:enable:var-naming
)

var DefaultLogFormat = JSONLogFormat

// ConvertToZapFormat converts and validated log format string.
func ConvertToZapFormat(format string) (string, error) {
	switch format {
	case ConsoleLogFormat:
		return ConsoleLogFormat, nil
	case JSONLogFormat:
		return JSONLogFormat, nil
	case "":
		return DefaultLogFormat, nil
	default:
		return "", fmt.Errorf("unknown log format: %s, supported values json, console", format)
	}
}
