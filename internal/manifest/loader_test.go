package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "manifest-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoadFromFile_Valid(t *testing.T) {
	content := `
version: "1"
containers:
  - name: web
    image: nginx:1.25
    ports:
      - "80:80"
    env:
      ENV: production
    restartPolicy: always
`
	path := writeTemp(t, content)
	m, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(m.Containers))
	}
	if m.Containers[0].Image != "nginx:1.25" {
		t.Errorf("expected image nginx:1.25, got %s", m.Containers[0].Image)
	}
}

func TestLoadFromFile_MissingVersion(t *testing.T) {
	content := `containers:
  - name: app
    image: myapp:latest
`
	path := writeTemp(t, content)
	_, err := LoadFromFile(path)
	if err == nil {
		t.Fatal("expected error for missing version, got nil")
	}
}

func TestLoadFromFile_DuplicateName(t *testing.T) {
	content := `
version: "1"
containers:
  - name: web
    image: nginx:latest
  - name: web
    image: apache:latest
`
	path := writeTemp(t, content)
	_, err := LoadFromFile(path)
	if err == nil {
		t.Fatal("expected error for duplicate container name, got nil")
	}
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := LoadFromFile(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
