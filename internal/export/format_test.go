package export

import (
	"testing"
)

func TestParseFormat_Valid(t *testing.T) {
	cases := []struct {
		input string
		want  Format
	}{
		{"json", FormatJSON},
		{"markdown", FormatMarkdown},
		{"html", FormatHTML},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseFormat(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseFormat_Invalid(t *testing.T) {
	_, err := ParseFormat("xml")
	if err == nil {
		t.Error("expected error for unknown format")
	}
}

func TestSupportedFormats(t *testing.T) {
	formats := SupportedFormats()
	if len(formats) != 3 {
		t.Errorf("expected 3 formats, got %d", len(formats))
	}
	seen := map[string]bool{}
	for _, f := range formats {
		seen[f] = true
	}
	for _, want := range []string{"json", "markdown", "html"} {
		if !seen[want] {
			t.Errorf("missing format %q in SupportedFormats", want)
		}
	}
}
