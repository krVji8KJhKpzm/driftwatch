package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/driftwatch/internal/drift"
)

// EventType classifies the kind of audit event.
type EventType string

const (
	EventDriftDetected  EventType = "drift_detected"
	EventDriftResolved  EventType = "drift_resolved"
	EventScanCompleted  EventType = "scan_completed"
)

// Event represents a single audit log entry.
type Event struct {
	Timestamp     time.Time `json:"timestamp"`
	Type          EventType `json:"type"`
	ContainerName string    `json:"container_name"`
	Details       string    `json:"details,omitempty"`
}

// Logger writes audit events to a file.
type Logger struct {
	path string
}

// NewLogger creates a Logger that appends events to the given file path.
func NewLogger(path string) *Logger {
	return &Logger{path: path}
}

// Record appends an audit event to the log file.
func (l *Logger) Record(evt Event) error {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("audit: open %s: %w", l.path, err)
	}
	defer f.Close()

	data, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("audit: marshal event: %w", err)
	}
	_, err = fmt.Fprintf(f, "%s\n", data)
	return err
}

// FromResults generates and records audit events for a slice of drift results.
func (l *Logger) FromResults(results []drift.Result) error {
	for _, r := range results {
		evt := Event{
			Timestamp:     time.Now().UTC(),
			ContainerName: r.ContainerName,
		}
		if r.Drifted {
			evt.Type = EventDriftDetected
			evt.Details = fmt.Sprintf("%d field(s) drifted", len(r.Fields))
		} else {
			evt.Type = EventScanCompleted
		}
		if err := l.Record(evt); err != nil {
			return err
		}
	}
	return nil
}

// LoadEvents reads all audit events from the given file path.
func LoadEvents(path string) ([]Event, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("audit: read %s: %w", path, err)
	}

	var events []Event
	for _, line := range splitLines(data) {
		if len(line) == 0 {
			continue
		}
		var evt Event
		if err := json.Unmarshal(line, &evt); err != nil {
			return nil, fmt.Errorf("audit: parse line: %w", err)
		}
		events = append(events, evt)
	}
	return events, nil
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
