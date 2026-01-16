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

## Status

🚧 **Work in Progress** - Initial development in progress.

## Installation

```bash
go get github.com/pfrederiksen/configdiff
```

## Quick Start

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
