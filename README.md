# Prepush

[![Build Status][build-status-svg]][build-status-url]
[![Lint Status][lint-status-svg]][lint-status-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![License][license-svg]][license-url]

**Multi-language pre-push hook for Git repositories.**

`prepush` automatically detects languages in your repository and runs appropriate checks before you push. It supports monorepos with multiple languages.

## Features

- **Auto-detection**: Detects Go, TypeScript, JavaScript, Python, Rust, Swift
- **Monorepo support**: Handles repositories with multiple languages
- **Configurable**: Optional `.prepush.yaml` for customization
- **Fast**: Only runs checks for detected languages

## Installation

```bash
go install github.com/grokify/prepush@latest
```

## Usage

### Basic Usage

```bash
# Run in current directory
prepush

# Run on specific directory
prepush /path/to/repo

# Verbose output
prepush --verbose

# Skip specific checks
prepush --no-test
prepush --no-lint
prepush --no-format

# Show coverage (Go only)
prepush --coverage
```

### As Git Hook

To run prepush automatically before every `git push`:

**Option 1: Create hook script**

```bash
# Create the hook
cat > .git/hooks/pre-push << 'EOF'
#!/bin/bash
exec prepush
EOF

# Make it executable
chmod +x .git/hooks/pre-push
```

**Option 2: Symlink (simpler)**

```bash
ln -sf $(which prepush) .git/hooks/pre-push
```

**Option 3: Shared hooks (team setup)**

Since `.git/hooks/` isn't tracked by git, teams can use a shared hooks directory:

```bash
# Create tracked hooks directory
mkdir -p .githooks
cat > .githooks/pre-push << 'EOF'
#!/bin/bash
exec prepush
EOF
chmod +x .githooks/pre-push

# Configure git to use it (each team member runs this once)
git config core.hooksPath .githooks
```

**Bypassing the hook**

When needed (e.g., WIP branches), bypass with:

```bash
git push --no-verify
```

## Supported Languages

| Language | Detection | Checks |
|----------|-----------|--------|
| **Go** | `go.mod` | `go build`, `go mod tidy`, `gofmt`, `golangci-lint`, `go test`, local replace, untracked refs |
| **TypeScript** | `package.json` + `tsconfig.json` | `eslint`, `prettier`, `tsc --noEmit`, `npm test` |
| **JavaScript** | `package.json` | `eslint`, `prettier`, `npm test` |
| **Python** | `pyproject.toml`, `setup.py`, `requirements.txt` | Coming soon |
| **Rust** | `Cargo.toml` | Coming soon |
| **Swift** | `Package.swift` | Coming soon |

### Go Checks Detail

| Check | Type | Description |
|-------|------|-------------|
| no local replace | Hard | Fails if go.mod has local replace directives |
| mod tidy | Hard | Fails if go.mod/go.sum need updating |
| build | Hard | Fails if project doesn't compile |
| gofmt | Hard | Fails if code isn't formatted |
| golangci-lint | Hard | Fails if linter reports issues |
| tests | Hard | Fails if tests fail |
| untracked refs | Soft | Warns if tracked files reference untracked files |
| coverage | Soft | Reports coverage (requires `gocoverbadge`) |

## Configuration

Create `.prepush.yaml` in your repository root:

```yaml
# Global settings
verbose: false

# Language-specific settings
languages:
  go:
    enabled: true
    test: true
    lint: true
    format: true
    coverage: false
    exclude_coverage: "cmd"  # directories to exclude from coverage

  typescript:
    enabled: true
    paths: ["frontend/"]  # specific paths (empty = auto-detect)
    test: true
    lint: true
    format: true

  javascript:
    enabled: false  # disable for this repo
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable/disable language checks |
| `paths` | []string | auto | Specific paths to check |
| `test` | bool | `true` | Run tests |
| `lint` | bool | `true` | Run linter |
| `format` | bool | `true` | Check formatting |
| `coverage` | bool | `false` | Show coverage (Go only) |

## Examples

### Go Project

```
$ prepush
=== Pre-push Checks ===

Detecting languages...
  Found: go in .

Running Go checks...

=== Summary ===
✓ Go: no local replace directives
✓ Go: mod tidy
✓ Go: build
✓ Go: gofmt
✓ Go: golangci-lint
✓ Go: tests
✓ Go: untracked references

Passed: 7, Failed: 0, Skipped: 0

All pre-push checks passed!
```

### With Warnings

```
$ prepush
=== Pre-push Checks ===

Detecting languages...
  Found: go in .

Running Go checks...

=== Summary ===
✓ Go: no local replace directives
✓ Go: mod tidy
✓ Go: build
✓ Go: gofmt
✓ Go: golangci-lint
✓ Go: tests
⚠ Go: untracked references (warning)
  main.go may reference untracked utils.go

Passed: 6, Failed: 0, Skipped: 0, Warnings: 1

Pre-push checks passed with warnings.
```

### Monorepo (Go + TypeScript)

```
$ prepush
=== Pre-push Checks ===

Detecting languages...
  Found: go in ./backend
  Found: typescript in ./frontend

Running Go checks...
Running TypeScript checks...

=== Summary ===
✓ Go: no local replace directives
✓ Go: gofmt
✓ Go: golangci-lint
✓ Go: tests
✓ TypeScript: eslint
✓ TypeScript: prettier
✓ TypeScript: type check
✓ TypeScript: tests

Passed: 8, Failed: 0, Skipped: 0

All pre-push checks passed!
```

## Related Tools

- [gocoverbadge](https://github.com/grokify/mogo/tree/main/testing/coverbadge) - Generate coverage badges for Go projects

## License

MIT License - see [LICENSE](LICENSE) for details.

 [build-status-svg]: https://github.com/grokify/prepush/actions/workflows/ci.yaml/badge.svg?branch=main
 [build-status-url]: https://github.com/grokify/prepush/actions/workflows/ci.yaml
 [lint-status-svg]: https://github.com/grokify/prepush/actions/workflows/lint.yaml/badge.svg?branch=main
 [lint-status-url]: https://github.com/grokify/prepush/actions/workflows/lint.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/grokify/prepush
 [goreport-url]: https://goreportcard.com/report/github.com/grokify/prepush
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/grokify/prepush
 [docs-godoc-url]: https://pkg.go.dev/github.com/grokify/prepush
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/grokify/prepush/blob/master/LICENSE
 [used-by-svg]: https://sourcegraph.com/github.com/grokify/prepush/-/badge.svg
 [used-by-url]: https://sourcegraph.com/github.com/grokify/prepush?badge
