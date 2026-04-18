# driftwatch

CLI tool that detects configuration drift between running containers and their source manifests.

---

## Installation

```bash
go install github.com/yourusername/driftwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/driftwatch.git
cd driftwatch
go build -o driftwatch .
```

## Usage

Point driftwatch at a manifest file and let it compare against your running containers:

```bash
# Check drift against a Kubernetes manifest
driftwatch check --manifest deployment.yaml --namespace production

# Watch for drift continuously
driftwatch watch --manifest docker-compose.yml --interval 30s

# Output results as JSON
driftwatch check --manifest deployment.yaml --output json
```

Example output:

```
[DRIFT DETECTED] container: api-server
  expected image: myapp:v1.2.0
  running image:  myapp:v1.1.9

  expected replicas: 3
  running replicas:  2
```

## Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--manifest` | Path to source manifest file | required |
| `--namespace` | Kubernetes namespace to inspect | `default` |
| `--interval` | Poll interval for watch mode | `60s` |
| `--output` | Output format (`text`, `json`) | `text` |

## Contributing

Pull requests are welcome. Please open an issue first to discuss any significant changes.

## License

[MIT](LICENSE)