---
name: commit-classification
description: Classifies commits according to conventional commits specification and maps them to changelog categories. Use when categorizing changes for changelogs or understanding commit types.
triggers: [classify, commit type, changelog category, conventional commit]
dependencies: [git, schangelog]
---

# Commit Classification

Classifies commits according to conventional commits specification and maps them to changelog categories. Use when categorizing changes for changelogs or understanding commit types.

## Instructions

Classify git commits according to the Conventional Commits specification (v1.0.0).

## Commit Types and Changelog Mapping

| Type | Changelog Category |
|------|-------------------|
| feat | Added |
| fix | Fixed |
| docs | Documentation |
| refactor | Changed |
| perf | Changed |
| revert | Removed |

## Breaking Changes

Indicated by:
- `!` after type: `feat!: remove API`
- `BREAKING CHANGE:` in footer

## Classification Process

```bash
schangelog parse-commits --since=<tag>
```

## Output Format

```
Commit: <hash>
Subject: <subject>
Type: <type>
Scope: <scope>
Breaking: yes/no
Category: <changelog category>
```

