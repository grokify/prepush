---
description: Generate or update CHANGELOG.md for version $ARGUMENTS
---

# Changelog

Generate or update the CHANGELOG.md file using schangelog.

## Usage

```
/release-agent:changelog v1.2.3
```

## Process

1. Parse commits since last tag using `schangelog parse-commits`
2. Classify commits by type (feat, fix, docs, etc.)
3. Generate changelog entry for the specified version
4. Update CHANGELOG.md (or create if it doesn't exist)

## Prerequisites

- `schangelog` must be installed
- Git repository with conventional commits
- Existing tags for version history

## Output Format

Generates changelog in [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
## [1.2.3] - 2024-01-15

### Added
- feat: new feature description

### Fixed
- fix: bug fix description

### Changed
- refactor: change description
```

## Options

Preview without making changes:
```bash
schangelog parse-commits --since=$(git describe --tags --abbrev=0)
```

Generate to specific file:
```bash
schangelog generate CHANGELOG.json -o CHANGELOG.md
```
