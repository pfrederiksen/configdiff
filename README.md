# configdiff

Semantic, human-grade diffs for YAML/JSON/HCL configuration files.

## Overview

`configdiff` provides intelligent semantic diffing for configuration files that goes beyond simple line-based comparison. It understands the structure of your configuration and can:

- Normalize different formats (YAML, JSON, HCL) into a common representation
- Apply customizable rules for semantic comparison
- Ignore specific paths or treat arrays as sets
- Handle type coercions (e.g., `"1"` vs `1`, `"true"` vs `true`)
- Generate both machine-readable patches and human-friendly reports

Perfect for GitOps reviews, CI checks, configuration drift detection, and any scenario where you need to understand what actually changed in your config files.

## Installation

### CLI Tool

```bash
# Homebrew (macOS/Linux)
brew install pfrederiksen/tap/configdiff

# Or download binaries from GitHub releases
# https://github.com/pfrederiksen/configdiff/releases
```

### Go Library

```bash
go get github.com/pfrederiksen/configdiff
```

## Quick Start

### CLI Usage

```bash
# Basic comparison
configdiff old.yaml new.yaml

# Compare with stdin
kubectl get deploy myapp -o yaml | configdiff old.yaml -

# Different output formats
configdiff old.yaml new.yaml -o compact
configdiff old.yaml new.yaml -o json
configdiff old.yaml new.yaml -o patch

# Ignore specific paths
configdiff old.yaml new.yaml -i /metadata/generation -i /status/*

# Array-as-set comparison
configdiff old.yaml new.yaml --array-key /spec/containers=name

# Exit code mode for CI
if configdiff old.yaml new.yaml --exit-code; then
  echo "No changes detected"
fi
```

### Library Usage

```go
package main

import (
    "fmt"
    "github.com/pfrederiksen/configdiff"
)

func main() {
    oldYAML := []byte(`
name: myapp
replicas: 3
image: nginx:1.19
`)

    newYAML := []byte(`
name: myapp
replicas: 5
image: nginx:1.20
env: production
`)

    result, err := configdiff.DiffYAML(oldYAML, newYAML, configdiff.Options{})
    if err != nil {
        panic(err)
    }

    // Human-friendly report
    fmt.Println(result.Report)
    // Output:
    // Summary: +1 added, ~2 modified (3 total)
    //
    // Changes:
    //   + /env = "production"
    //
    //   ~ /image: "nginx:1.19" → "nginx:1.20"
    //
    //   ~ /replicas: 3 → 5

    // Machine-readable patch
    patchJSON, _ := result.Patch.ToJSONIndent()
    fmt.Println(string(patchJSON))
}
```

## CLI Reference

### Flags

```
Format Options:
  -f, --format string          Input format (yaml, json, auto) (default "auto")
      --old-format string      Old file format override
      --new-format string      New file format override

Diff Options:
  -i, --ignore strings         Paths to ignore (can be repeated)
      --array-key strings      Array paths to key fields (format: path=key)
      --numeric-strings        Coerce numeric strings to numbers
      --bool-strings           Coerce bool strings to booleans
      --stable-order           Sort output deterministically (default true)

Output Options:
  -o, --output string          Output format (report, compact, json, patch) (default "report")
      --no-color               Disable colored output
      --max-value-length int   Truncate values longer than N chars (default 80)
  -q, --quiet                  Quiet mode (no output)
      --exit-code              Exit with code 1 if differences found

Other:
  -h, --help                   Help for configdiff
  -v, --version                Version information
```

### Output Formats

- **report** (default): Detailed human-friendly report with values
- **compact**: Summary with paths only
- **json**: JSON-serialized changes array
- **patch**: JSON Patch (RFC 6902) format

### Exit Codes

- `0`: Success (no differences, or differences displayed)
- `1`: Differences found (when using `--exit-code`)
- `1`: Error occurred

## Features

### Normalized Tree Representation

All configuration formats are parsed into a normalized tree structure with explicit node types:
- Null, Bool, Number, String
- Object (key-value mappings)
- Array (ordered lists)

### Customizable Diff Rules

Configure how diffs are computed:

```go
opts := configdiff.Options{
    // Ignore specific paths
    IgnorePaths: []string{
        "metadata.creationTimestamp",
        "status.*",
    },

    // Treat arrays as sets keyed by a field
    ArraySetKeys: map[string]string{
        "spec.containers": "name",
        "spec.volumes": "name",
    },

    // Enable type coercions
    Coercions: configdiff.Coercions{
        NumericStrings: true,
        BoolStrings: true,
    },

    // Stable ordering for deterministic output
    StableOrder: true,
}
```

### Multiple Output Formats

**Machine-readable patches** (JSON Patch-like):
```json
{
  "operations": [
    {
      "op": "add",
      "path": "/env",
      "value": "production"
    },
    {
      "op": "replace",
      "path": "/image",
      "value": "nginx:1.20"
    },
    {
      "op": "replace",
      "path": "/replicas",
      "value": 5
    }
  ]
}
```

**Pretty reports** with configurable verbosity:
```go
// Detailed report (default)
report.GenerateDetailed(changes)
// Summary: +1 added, ~2 modified (3 total)
//
// Changes:
//   + /env = "production"
//   ~ /image: "nginx:1.19" → "nginx:1.20"
//   ~ /replicas: 3 → 5

// Compact report (paths only)
report.GenerateCompact(changes)
// Summary: +1 added, ~2 modified (3 total)
// Changes:
//   + /env
//   ~ /image
//   ~ /replicas

// Custom formatting
report.Generate(changes, report.Options{
    Compact: false,
    ShowValues: true,
    MaxValueLength: 50,  // Truncate long values
})
```

## Examples

### Ignore Specific Paths

Useful for ignoring timestamps, auto-generated fields, or status information:

```go
opts := configdiff.Options{
    IgnorePaths: []string{
        "/metadata/creationTimestamp",
        "/metadata/generation",
        "/status",           // Exact match
        "/status/*",         // Wildcard: ignores all fields under /status
    },
}

result, _ := configdiff.DiffYAML(oldK8s, newK8s, opts)
```

### Array-as-Set Comparison

Compare arrays by a key field instead of position:

```go
oldYAML := []byte(`
spec:
  containers:
    - name: nginx
      image: nginx:1.19
    - name: sidecar
      image: busybox:latest
`)

newYAML := []byte(`
spec:
  containers:
    - name: sidecar
      image: busybox:1.36    # Reordered + changed
    - name: nginx
      image: nginx:1.20      # Changed
`)

opts := configdiff.Options{
    ArraySetKeys: map[string]string{
        "/spec/containers": "name",  // Match containers by "name" field
    },
}

result, _ := configdiff.DiffYAML(oldYAML, newYAML, opts)
// Output:
// Summary: ~2 modified (2 total)
//
// Changes:
//   ~ /spec/containers[name=nginx]/image: "nginx:1.19" → "nginx:1.20"
//   ~ /spec/containers[name=sidecar]/image: "busybox:latest" → "busybox:1.36"
```

### Type Coercions

Handle semantic equivalence across type boundaries:

```go
jsonConfig := []byte(`{"replicas": 3, "enabled": true}`)
yamlConfig := []byte(`
replicas: "3"     # String in YAML
enabled: "true"   # String in YAML
`)

opts := configdiff.Options{
    Coercions: configdiff.Coercions{
        NumericStrings: true,  // "3" == 3
        BoolStrings: true,     // "true" == true
    },
}

result, _ := configdiff.DiffBytes(jsonConfig, "json", yamlConfig, "yaml", opts)
// No differences detected due to coercion
```

### Cross-Format Comparison

Compare YAML and JSON representations:

```go
yamlConfig := []byte(`
database:
  host: localhost
  port: 5432
`)

jsonConfig := []byte(`{
  "database": {
    "host": "localhost",
    "port": 5432
  }
}`)

result, _ := configdiff.DiffBytes(yamlConfig, "yaml", jsonConfig, "json", configdiff.Options{})
// No differences - semantically identical
```

### Kubernetes Deployment Diff

Real-world example comparing Kubernetes deployments:

```go
package main

import (
    "fmt"
    "github.com/pfrederiksen/configdiff"
)

func main() {
    oldDeploy := []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  generation: 1
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: app
        image: myapp:v1.0
        resources:
          limits:
            memory: 512Mi
`)

    newDeploy := []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  generation: 2
spec:
  replicas: 5
  template:
    spec:
      containers:
      - name: app
        image: myapp:v1.1
        resources:
          limits:
            memory: 1Gi
      - name: sidecar
        image: envoy:v1.20
`)

    opts := configdiff.Options{
        IgnorePaths: []string{
            "/metadata/generation",  // Auto-incremented
        },
        ArraySetKeys: map[string]string{
            "/spec/template/spec/containers": "name",
        },
        StableOrder: true,
    }

    result, _ := configdiff.DiffYAML(oldDeploy, newDeploy, opts)
    fmt.Println(result.Report)
    // Output:
    // Summary: +1 added, ~3 modified (4 total)
    //
    // Changes:
    //   + /spec/template/spec/containers[name=sidecar] = {...} (2 keys)
    //
    //   ~ /spec/replicas: 3 → 5
    //
    //   ~ /spec/template/spec/containers[name=app]/image: "myapp:v1.0" → "myapp:v1.1"
    //
    //   ~ /spec/template/spec/containers[name=app]/resources/limits/memory: "512Mi" → "1Gi"
}
```

## API Reference

### Main Functions

```go
// DiffBytes compares two configuration byte slices
func DiffBytes(a []byte, aFormat string, b []byte, bFormat string, opts Options) (*Result, error)

// DiffYAML is a convenience function for YAML-only comparison
func DiffYAML(a, b []byte, opts Options) (*Result, error)

// DiffJSON is a convenience function for JSON-only comparison
func DiffJSON(a, b []byte, opts Options) (*Result, error)

// DiffTrees compares pre-parsed tree nodes
func DiffTrees(a, b *tree.Node, opts Options) (*Result, error)
```

### Options

```go
type Options struct {
    // IgnorePaths: List of paths to ignore during comparison
    // Supports wildcards: "/status/*" matches all fields under /status
    IgnorePaths []string

    // ArraySetKeys: Map of array paths to key fields
    // Treats arrays as sets, matching elements by the specified field
    // Example: map[string]string{"/spec/containers": "name"}
    ArraySetKeys map[string]string

    // Coercions: Type coercion rules for semantic comparison
    Coercions Coercions

    // StableOrder: Sort changes deterministically for reproducible output
    StableOrder bool
}

type Coercions struct {
    // NumericStrings: Treat numeric strings as numbers ("42" == 42)
    NumericStrings bool

    // BoolStrings: Treat bool strings as booleans ("true" == true)
    BoolStrings bool
}
```

### Result

```go
type Result struct {
    // Changes: List of detected changes
    Changes []Change

    // Patch: Machine-readable patch operations
    Patch *Patch

    // Report: Human-friendly formatted report
    Report string
}

type Change struct {
    Type     ChangeType  // Add, Remove, Modify, Move
    Path     string      // JSON Pointer-like path
    OldValue *tree.Node  // Previous value (nil for Add)
    NewValue *tree.Node  // New value (nil for Remove)
}
```

### Report Generation

```go
// Generate creates a report with custom options
func Generate(changes []Change, opts Options) string

// GenerateDetailed creates a detailed report with values
func GenerateDetailed(changes []Change) string

// GenerateCompact creates a compact report with paths only
func GenerateCompact(changes []Change) string

type Options struct {
    Compact        bool  // If true, only show paths
    ShowValues     bool  // If true, include old/new values
    MaxValueLength int   // Truncate values longer than this (0 = no limit)
}
```

## Use Cases

- **GitOps Reviews**: Understand exactly what changed in infrastructure configs
- **CI/CD Checks**: Validate configuration changes before deployment
- **Drift Detection**: Compare actual vs desired state in deployed systems
- **Configuration Management**: Track changes across environments
- **Multi-Format Comparison**: Compare YAML and JSON representations of the same config

## Testing

The project maintains high test coverage (>80%) with comprehensive test suites:

- **Unit Tests**: Table-driven tests for all core functionality
- **Golden Tests**: Reference output files in `testdata/` for report formatting validation
- **Integration Tests**: End-to-end scenarios covering real-world use cases
- **CI/CD**: Automated testing on multiple Go versions (1.21, 1.22) with coverage enforcement

Run tests:
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detector
go test -race ./...

# Update golden test files
go test ./report -update
```

## Project Status

**Production Ready** - All core features implemented and tested:

- [x] Repository setup with CI/CD
- [x] Tree package with normalized representation
- [x] YAML/JSON parsing with format detection
- [x] Semantic diff engine with customizable rules
- [x] JSON Patch-like operations
- [x] Human-friendly report generation
- [x] Comprehensive test coverage (84.8%)
- [x] Full API documentation and examples

### Future Enhancements

- HCL format support (experimental)
- Additional coercion rules
- Performance optimizations for very large configs
- CLI tool for command-line usage

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
