---
name: commit-classification
description: Classifies commits according to conventional commits specification and maps them to changelog categories. Use when categorizing changes for changelogs or understanding commit types.
allowed-tools: Read, Grep, Bash
model: haiku
user-invocable: true
---

# Commit Classification Skill

Classify git commits according to the Conventional Commits specification (v1.0.0) and map them to appropriate changelog categories.

## Conventional Commits Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

## Commit Types and Changelog Mapping

| Type | Description | Changelog Category |
|------|-------------|-------------------|
| `feat` | New feature | Added |
| `fix` | Bug fix | Fixed |
| `docs` | Documentation only | Documentation |
| `style` | Formatting, no code change | (usually omitted) |
| `refactor` | Code change, no feature/fix | Changed |
| `perf` | Performance improvement | Changed |
| `test` | Adding/correcting tests | (usually omitted) |
| `build` | Build system changes | Changed |
| `ci` | CI configuration | (usually omitted) |
| `chore` | Maintenance tasks | (usually omitted) |
| `revert` | Reverts a commit | Removed |

## Breaking Changes

Breaking changes are indicated by:
1. `!` after type/scope: `feat!: remove deprecated API`
2. `BREAKING CHANGE:` footer in commit body
3. `BREAKING-CHANGE:` footer (alternative format)

Breaking changes always go to **Changed** with special notation.

## Scope Examples

Common scopes for release-agent:
- `feat(cli):` - CLI changes
- `fix(workflow):` - Workflow fixes
- `docs(readme):` - README updates
- `refactor(actions):` - Action refactoring

## Classification Process

### Using schangelog (Recommended)
```bash
# Parses commits and suggests categories
schangelog parse-commits --since=<tag>
```

### Manual Classification
```bash
# Get commit subjects
git log <tag>..HEAD --format="%s"

# Parse each commit:
# 1. Extract type (before colon)
# 2. Check for ! (breaking)
# 3. Extract scope (in parentheses)
# 4. Map to changelog category
```

## Output Format

When classifying commits, output:
```
Commit: <hash>
Subject: <subject>
Type: <type>
Scope: <scope or none>
Breaking: <yes/no>
Category: <changelog category>
```

## Example Classifications

```
Commit: abc1234
Subject: feat(api): add new endpoint for releases
Type: feat
Scope: api
Breaking: no
Category: Added

Commit: def5678
Subject: fix!: correct version parsing
Type: fix
Scope: none
Breaking: yes
Category: Changed (Breaking)

Commit: ghi9012
Subject: docs: update installation guide
Type: docs
Scope: none
Breaking: no
Category: Documentation
```
