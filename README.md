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
`)

    newYAML := []byte(`
name: myapp
replicas: 5
`)

    result, err := configdiff.DiffBytes(oldYAML, "yaml", newYAML, "yaml", configdiff.Options{})
    if err != nil {
        panic(err)
    }

    fmt.Println(result.Report)
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

- **Machine-readable patches**: JSON Patch-like operations for programmatic consumption
- **Pretty reports**: Human-friendly, concise output with context

## Use Cases

- **GitOps Reviews**: Understand exactly what changed in infrastructure configs
- **CI/CD Checks**: Validate configuration changes before deployment
- **Drift Detection**: Compare actual vs desired state in deployed systems
- **Configuration Management**: Track changes across environments

## Development Status

This project is under active development. Current progress:

- [x] Repository setup
- [ ] Tree package implementation
- [ ] YAML/JSON parsing
- [ ] Diff engine
- [ ] Patch format
- [ ] Pretty reporting
- [ ] Comprehensive tests
- [ ] Full documentation

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
