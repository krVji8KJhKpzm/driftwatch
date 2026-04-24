package watch

import (
	"context"
	"log"
	"time"

	"github.com/driftwatch/internal/container"
	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/manifest"
	"github.com/driftwatch/internal/report"
)

// Config holds configuration for the watch loop.
type Config struct {
	ManifestPath string
	Interval     time.Duration
	Format       string
	Output       string
}

// Runner abstracts the watch loop for testing.
type Runner interface {
	Run(ctx context.Context, cfg Config) error
}

type defaultRunner struct {
	inspector container.Inspector
}

// NewRunner creates a new watch Runner using the default Docker inspector.
func NewRunner(inspector container.Inspector) Runner {
	return &defaultRunner{inspector: inspector}
}

// Run starts a polling loop that detects drift at the given interval.
func (r *defaultRunner) Run(ctx context.Context, cfg Config) error {
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	log.Printf("[watch] starting drift watch every %s", cfg.Interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("[watch] stopping")
			return ctx.Err()
		case <-ticker.C:
			if err := r.runOnce(cfg); err != nil {
				log.Printf("[watch] error during drift check: %v", err)
			}
		}
	}
}

func (r *defaultRunner) runOnce(cfg Config) error {
	entries, err := manifest.LoadFromFile(cfg.ManifestPath)
	if err != nil {
		return err
	}

	infos, err := r.inspector.Inspect(ctx(entries))
	if err != nil {
		return err
	}

	results := drift.Detect(entries, infos)
	return report.Write(results, cfg.Format, cfg.Output)
}

// ctx builds a minimal background context for inspector calls.
func ctx(_ interface{}) context.Context {
	return context.Background()
}
