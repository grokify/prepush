---
description: Analyze commits and suggest next semantic version
---

# Version Next

Analyze git history since the last tag and suggest the next semantic version based on conventional commits.

## Usage

```
/release-agent:version-next
```

## Process

1. Get current version from latest git tag
2. Parse commits since that tag
3. Classify commits by type
4. Determine version bump (major/minor/patch)
5. Suggest next version with reasoning

## Semantic Versioning Rules

Given `vMAJOR.MINOR.PATCH`:

| Change Type | Bump | Example |
|-------------|------|---------|
| Breaking change | MAJOR | v1.0.0 -> v2.0.0 |
| New feature (feat:) | MINOR | v1.0.0 -> v1.1.0 |
| Bug fix (fix:) | PATCH | v1.0.0 -> v1.0.1 |

## Example Output

```
=== Version Analysis ===

Current Version: v0.8.0
Commits Since Tag: 15

Breakdown:
  - feat: 3 (new features)
  - fix: 5 (bug fixes)
  - docs: 4 (documentation)
  - chore: 3 (maintenance)
  - BREAKING: 0

Suggested Next Version: v0.9.0

Reasoning: 3 new features warrant a minor version bump.
No breaking changes detected.
```

## Commands Used

```bash
# Get current version
git describe --tags --abbrev=0

# Analyze commits (token-efficient)
schangelog parse-commits --since=$(git describe --tags --abbrev=0)
```
