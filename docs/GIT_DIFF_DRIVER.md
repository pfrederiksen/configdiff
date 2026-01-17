# Git Diff Driver Integration

Configure git to automatically use `configdiff` for semantic diffs of configuration files.

## What is a Git Diff Driver?

A git diff driver is a custom tool that git uses to generate diffs for specific file types. Instead of showing line-by-line text diffs, configdiff provides semantic, structure-aware comparisons of YAML, JSON, and HCL files.

## Setup

### 1. Install configdiff

```bash
# Via Homebrew
brew install pfrederiksen/tap/configdiff

# Or download from releases
# https://github.com/pfrederiksen/configdiff/releases
```

### 2. Configure Git (Global)

Add the following to your `~/.gitconfig`:

```ini
[diff "configdiff"]
    command = configdiff --output git-diff
    textconv = configdiff --output git-diff
```

### 3. Configure Git Attributes (Per-Repository)

Create or edit `.gitattributes` in your repository root:

```gitattributes
# YAML files
*.yaml diff=configdiff
*.yml diff=configdiff

# JSON files
*.json diff=configdiff

# HCL files (Terraform, etc.)
*.hcl diff=configdiff
*.tf diff=configdiff

# Kubernetes manifests
*.k8s.yaml diff=configdiff
```

### 4. Test It

```bash
# Make a change to a config file
echo 'replicas: 5' >> deployment.yaml

# View the semantic diff
git diff deployment.yaml
```

You should see configdiff's semantic output instead of line-by-line diffs.

## Advanced Configuration

### Per-Repository Settings

For repository-specific settings, use `.git/config` instead of `~/.gitconfig`:

```bash
git config diff.configdiff.command "configdiff --output git-diff"
```

### Custom Options

You can pass additional flags to configdiff:

```ini
[diff "configdiff-nocolor"]
    command = configdiff --output git-diff --no-color

[diff "configdiff-ignore-metadata"]
    command = configdiff --output git-diff --ignore /metadata/*
```

Then in `.gitattributes`:

```gitattributes
deployment.yaml diff=configdiff-ignore-metadata
```

### Using with Other Output Formats

While `git-diff` format is recommended for git integration, you can use other formats:

```ini
# Statistics summary (like git diff --stat)
[diff "configdiff-stat"]
    command = configdiff --output stat

# Side-by-side comparison
[diff "configdiff-sidebyside"]
    command = configdiff --output side-by-side

# Detailed report
[diff "configdiff-report"]
    command = configdiff --output report
```

## Troubleshooting

### Diff doesn't show up

Check that:
1. `configdiff` is in your PATH: `which configdiff`
2. `.gitattributes` is committed and has correct patterns
3. Git config is set: `git config --get diff.configdiff.command`

### Wrong format displayed

Ensure you're using the correct output format in your git config:
```bash
git config diff.configdiff.command "configdiff --output git-diff"
```

### Permission denied

Make sure configdiff is executable:
```bash
chmod +x $(which configdiff)
```

## Examples

### Before (standard git diff)

```diff
diff --git a/config.yaml b/config.yaml
index 1234567..abcdefg 100644
--- a/config.yaml
+++ b/config.yaml
@@ -10,7 +10,7 @@ spec:
   containers:
   - name: app
-    image: nginx:1.19
+    image: nginx:1.20
     ports:
-  replicas: 2
+  replicas: 3
```

### After (configdiff)

```diff
diff --configdiff a/config.yaml b/config.yaml
--- a/config.yaml
+++ b/config.yaml
@@ /spec/containers @@
-/spec/containers[0]/image: "nginx:1.19"
+/spec/containers[0]/image: "nginx:1.20"
@@ /spec/replicas @@
-/spec/replicas: 2
+/spec/replicas: 3
```

## Benefits

- **Semantic understanding**: Shows what actually changed in the configuration structure
- **Ignore formatting**: YAML indentation changes don't create noise
- **Array intelligence**: Detects array element changes even if order differs
- **Type awareness**: Understands `"2"` vs `2` differences when relevant
- **Path-based**: Clear indication of what configuration path changed

## Uninstalling

To remove the git diff driver:

```bash
# Remove from git config
git config --global --unset diff.configdiff.command
git config --global --unset diff.configdiff.textconv

# Remove .gitattributes entries
# Edit .gitattributes and remove the diff=configdiff lines
```

## See Also

- [configdiff Documentation](../README.md)
- [Git Attributes Documentation](https://git-scm.com/docs/gitattributes#_defining_a_custom_diff_driver)
- [Output Format Reference](../README.md#output-formats)
