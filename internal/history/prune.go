package history

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// PruneOptions controls which entries are removed.
type PruneOptions struct {
	// KeepLast retains only the N most recent entries. Zero means no limit.
	KeepLast int
	// OlderThan removes entries whose timestamp is before this time. Zero value means no cutoff.
	OlderThan time.Time
}

// Prune removes history entries from path according to opts.
// It returns the number of entries removed.
func Prune(path string, opts PruneOptions) (int, error) {
	entries, err := Load(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("prune: load: %w", err)
	}

	original := len(entries)

	// Filter by age first.
	if !opts.OlderThan.IsZero() {
		filtered := entries[:0]
		for _, e := range entries {
			if !e.Timestamp.Before(opts.OlderThan) {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	// Keep only the N most recent.
	if opts.KeepLast > 0 && len(entries) > opts.KeepLast {
		// Entries are sorted ascending; keep the tail.
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timestamp.Before(entries[j].Timestamp)
		})
		entries = entries[len(entries)-opts.KeepLast:]
	}

	removed := original - len(entries)
	if removed == 0 {
		return 0, nil
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("prune: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return 0, fmt.Errorf("prune: write: %w", err)
	}
	return removed, nil
}
