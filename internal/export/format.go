package export

import "fmt"

// ParseFormat converts a string to a Format, returning an error for unknown values.
func ParseFormat(s string) (Format, error) {
	switch Format(s) {
	case FormatJSON, FormatMarkdown, FormatHTML:
		return Format(s), nil
	default:
		return "", fmt.Errorf("unknown export format %q: must be one of json, markdown, html", s)
	}
}

// SupportedFormats returns all recognised format strings.
func SupportedFormats() []string {
	return []string{
		string(FormatJSON),
		string(FormatMarkdown),
		string(FormatHTML),
	}
}
