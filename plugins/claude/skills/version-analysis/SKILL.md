---
name: version-analysis
description: Analyzes git history to determine semantic version bumps based on conventional commits. Use when determining next version, analyzing recent changes, or reviewing commit history for release preparation.
triggers: [version, semver, next version, version bump]
dependencies: [git, schangelog]
---

# Version Analysis

Analyzes git history to determine semantic version bumps based on conventional commits. Use when determining next version, analyzing recent changes, or reviewing commit history for release preparation.

## Instructions

Analyze git history and conventional commits to determine appropriate semantic version bumps.

## Process

1. Get current version: `git describe --tags --abbrev=0`
2. Analyze commits: `schangelog parse-commits --since=<tag>`
3. Apply semantic versioning rules:
   - MAJOR: Breaking changes
   - MINOR: New features (feat:)
   - PATCH: Bug fixes (fix:)

## Output Format

```
Current Version: vX.Y.Z
Commits Since Tag: N commits
Breakdown:
  - Features (feat): M
  - Fixes (fix): N
  - Breaking Changes: P
Suggested Version: vA.B.C
Reasoning: Explanation
```

