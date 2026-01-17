# GitHub Action

Use configdiff in your GitHub Actions workflows to compare configuration files and detect changes in pull requests.

## Quick Start

```yaml
name: Config Diff

on: [pull_request]

jobs:
  diff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Fetch full history for comparing branches

      - name: Compare configs
        uses: pfrederiksen/configdiff@v0.2.0
        with:
          old-file: config/production.yaml
          new-file: config/staging.yaml
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `old-file` | Path to old configuration file or directory | Yes | - |
| `new-file` | Path to new configuration file or directory | Yes | - |
| `format` | Input format (yaml, json, hcl, toml, auto) | No | auto |
| `output-format` | Output format (report, compact, json, patch, stat, side-by-side, git-diff) | No | report |
| `ignore-paths` | Comma-separated list of paths to ignore | No | '' |
| `array-keys` | Comma-separated list of array key specs | No | '' |
| `numeric-strings` | Coerce numeric strings to numbers | No | false |
| `bool-strings` | Coerce bool strings to booleans | No | false |
| `no-color` | Disable colored output | No | false |
| `exit-code` | Exit with code 1 if differences found | No | false |
| `recursive` | Recursively compare directories | No | false |

## Outputs

| Output | Description |
|--------|-------------|
| `has-changes` | Whether any changes were detected (true/false) |
| `diff-output` | The diff output text |

## Examples

### Compare Files in PR

```yaml
name: PR Config Diff

on:
  pull_request:
    paths:
      - 'config/**'

jobs:
  config-diff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Check config changes
        uses: pfrederiksen/configdiff@v0.2.0
        with:
          old-file: config/production.yaml
          new-file: config/staging.yaml
          output-format: report
```

### Fail on Differences

```yaml
- name: Verify configs match
  uses: pfrederiksen/configdiff@v0.2.0
  with:
    old-file: expected.yaml
    new-file: actual.yaml
    exit-code: 'true'  # Fail if differences found
```

### Compare Directories

```yaml
- name: Compare config directories
  uses: pfrederiksen/configdiff@v0.2.0
  with:
    old-file: ./config-v1
    new-file: ./config-v2
    recursive: 'true'
```

### Ignore Specific Paths

```yaml
- name: Compare with ignored paths
  uses: pfrederiksen/configdiff@v0.2.0
  with:
    old-file: old.yaml
    new-file: new.yaml
    ignore-paths: '/metadata/generation,/status/*'
```

### Custom Output Format

```yaml
- name: Get diff statistics
  uses: pfrederiksen/configdiff@v0.2.0
  with:
    old-file: old.yaml
    new-file: new.yaml
    output-format: stat
```

### Compare Against Base Branch

```yaml
name: Compare Against Main

on:
  pull_request:
    branches: [main]

jobs:
  diff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Checkout base branch file
        run: |
          git show origin/main:config.yaml > /tmp/base-config.yaml

      - name: Compare with PR changes
        uses: pfrederiksen/configdiff@v0.2.0
        with:
          old-file: /tmp/base-config.yaml
          new-file: config.yaml
```

### Post Diff as PR Comment

```yaml
name: Config Diff Comment

on: [pull_request]

jobs:
  diff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run configdiff
        id: diff
        uses: pfrederiksen/configdiff@v0.2.0
        with:
          old-file: config/base.yaml
          new-file: config/updated.yaml
        continue-on-error: true

      - name: Comment PR
        uses: actions/github-script@v7
        if: steps.diff.outputs.has-changes == 'true'
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '## Configuration Changes\n\n```\n${{ steps.diff.outputs.diff-output }}\n```'
            })
```

### Matrix Testing with Multiple Formats

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        format: [yaml, json, toml]
    steps:
      - uses: actions/checkout@v4

      - name: Compare ${{ matrix.format }} configs
        uses: pfrederiksen/configdiff@v0.2.0
        with:
          old-file: test/old.${{ matrix.format }}
          new-file: test/new.${{ matrix.format }}
```

### Validate Kubernetes Manifests

```yaml
name: Validate K8s Manifests

on:
  pull_request:
    paths:
      - 'k8s/**'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Compare deployment configs
        uses: pfrederiksen/configdiff@v0.2.0
        with:
          old-file: k8s/production
          new-file: k8s/staging
          recursive: 'true'
          ignore-paths: '/metadata/creationTimestamp,/metadata/generation,/status/*'
          array-keys: '/spec/containers=name,/spec/volumes=name'
```

### Semantic Release Integration

```yaml
name: Release

on:
  push:
    branches: [main]

jobs:
  check-config:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 2

      - name: Get previous config
        run: git show HEAD~1:config.yaml > /tmp/prev-config.yaml

      - name: Check for breaking changes
        id: diff
        uses: pfrederiksen/configdiff@v0.2.0
        with:
          old-file: /tmp/prev-config.yaml
          new-file: config.yaml
          output-format: json

      - name: Determine version bump
        run: |
          # Parse JSON diff output to determine if breaking changes
          # Set RELEASE_TYPE=major if breaking, minor if new features, patch otherwise
          echo "RELEASE_TYPE=minor" >> $GITHUB_ENV
```

## Advanced Usage

### Using with Specific Version Tag

```yaml
uses: pfrederiksen/configdiff@v0.2.0
```

### Using with Latest

```yaml
uses: pfrederiksen/configdiff@main
```

### Using with Commit SHA (most secure)

```yaml
uses: pfrederiksen/configdiff@abc123...
```

## Troubleshooting

### Action Fails to Find Files

Ensure you've checked out the repository first:

```yaml
- uses: actions/checkout@v4
- uses: pfrederiksen/configdiff@v0.2.0
  with:
    old-file: path/to/file
    new-file: path/to/other/file
```

### Comparing Files from Different Branches

Use `git show` or checkout the other branch:

```yaml
- run: |
    git fetch origin
    git show origin/main:config.yaml > /tmp/main-config.yaml
- uses: pfrederiksen/configdiff@v0.2.0
  with:
    old-file: /tmp/main-config.yaml
    new-file: config.yaml
```

### Permission Denied on Docker

The action uses the published Docker image. If you encounter permission issues, ensure the files are readable:

```yaml
- run: chmod -R +r ./config
- uses: pfrederiksen/configdiff@v0.2.0
  with:
    old-file: ./config/old
    new-file: ./config/new
```

## See Also

- [configdiff CLI Documentation](../README.md)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub Actions Marketplace](https://github.com/marketplace)
