---
name: version-analysis
description: Analyzes git history to determine semantic version bumps based on conventional commits. Use when determining next version, analyzing recent changes, or reviewing commit history for release preparation.
allowed-tools: Read, Grep, Bash
model: sonnet
user-invocable: true
---

# Version Analysis Skill

Analyze git history and conventional commits to determine appropriate semantic version bumps.

## Process

### 1. Get Current Version
```bash
git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"
```

### 2. Analyze Commits (using schangelog for efficiency)
```bash
# TOON format is ~8x more token-efficient
schangelog parse-commits --since=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
```

### 3. Determine Version Bump

Based on conventional commits:

| Prefix | Bump | Example |
|--------|------|---------|
| `feat:` | Minor | New feature added |
| `fix:` | Patch | Bug fix |
| `BREAKING CHANGE:` | Major | Breaking API change |
| `feat!:` or `fix!:` | Major | Breaking change shorthand |
| `docs:`, `style:`, `refactor:`, `test:`, `chore:` | Patch | Non-functional changes |

### 4. Output Format

Provide analysis in this structure:
```
Current Version: vX.Y.Z
Commits Since Tag: N commits
Breakdown:
  - Features (feat): M
  - Fixes (fix): N
  - Breaking Changes: P
  - Other: Q
Suggested Version: vA.B.C
Reasoning: Specific explanation of why this version bump
```

## Semantic Versioning Rules

Given version `vMAJOR.MINOR.PATCH`:

- **MAJOR**: Incompatible API changes (breaking changes)
- **MINOR**: New functionality (backwards-compatible)
- **PATCH**: Bug fixes (backwards-compatible)

## Pre-release Versions

For pre-release versions:
- `v1.0.0-alpha.1` - Alpha release
- `v1.0.0-beta.1` - Beta release
- `v1.0.0-rc.1` - Release candidate

## Example Analysis

```
Current Version: v0.7.0
Commits Since Tag: 12 commits
Breakdown:
  - Features (feat): 3
  - Fixes (fix): 2
  - Breaking Changes: 0
  - Other (docs, chore): 7
Suggested Version: v0.8.0
Reasoning: 3 new features warrant a minor version bump. No breaking changes detected.
```
