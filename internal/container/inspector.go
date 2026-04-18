package container

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// ContainerInfo holds relevant runtime info for a container.
type ContainerInfo struct {
	ID      string            `json:"Id"`
	Name    string            `json:"-"`
	Image   string            `json:"-"`
	Env     map[string]string `json:"-"`
	Labels  map[string]string `json:"Labels"`
	RawEnv  []string          `json:"-"`
}

type dockerInspect struct {
	ID     string `json:"Id"`
	Name   string `json:"Name"`
	Config struct {
		Image  string            `json:"Image"`
		Env    []string          `json:"Env"`
		Labels map[string]string `json:"Labels"`
	} `json:"Config"`
}

// Runner abstracts command execution for testability.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// DefaultRunner executes real OS commands.
type DefaultRunner struct{}

func (r DefaultRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}

// Inspector fetches container runtime state.
type Inspector struct {
	Runner Runner
}

// NewInspector returns an Inspector using the real Docker CLI.
func NewInspector() *Inspector {
	return &Inspector{Runner: DefaultRunner{}}
}

// Inspect returns ContainerInfo for the given container name or ID.
func (i *Inspector) Inspect(ctx context.Context, containerName string) (*ContainerInfo, error) {
	out, err := i.Runner.Run(ctx, "docker", "inspect", containerName)
	if err != nil {
		return nil, fmt.Errorf("docker inspect %q: %w", containerName, err)
	}

	var results []dockerInspect
	if err := json.Unmarshal(out, &results); err != nil {
		return nil, fmt.Errorf("parse inspect output: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no container found: %q", containerName)
	}

	d := results[0]
	info := &ContainerInfo{
		ID:     d.ID,
		Name:   d.Name,
		Image:  d.Config.Image,
		Labels: d.Config.Labels,
		RawEnv: d.Config.Env,
		Env:    parseEnv(d.Config.Env),
	}
	return info, nil
}

func parseEnv(raw []string) map[string]string {
	m := make(map[string]string, len(raw))
	for _, e := range raw {
		for j := 0; j < len(e); j++ {
			if e[j] == '=' {
				m[e[:j]] = e[j+1:]
				break
			}
		}
	}
	return m
}
