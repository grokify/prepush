# Release Agent - Kiro CLI Plugin

Kiro CLI custom agents for automated software releases.

## Installation

Copy agent files to your Kiro agents directory:

```bash
# Copy all agents
cp plugins/kiro/agents/*.json ~/.kiro/agents/

# Or copy individual agents
cp plugins/kiro/agents/release-coordinator.json ~/.kiro/agents/
```

Optionally, copy steering files to your project:

```bash
mkdir -p .kiro/steering
cp plugins/kiro/steering/*.md .kiro/steering/
```

## Usage

### Start with Release Coordinator

```bash
kiro-cli --agent release-coordinator
```

### Switch Agents Mid-Session

```
> /agent swap release-qa
```

### Spawn Sub-Agents

The release coordinator can spawn specialized sub-agents:

```
> Use the release-qa agent to verify test coverage
> Use the release-security agent to check for vulnerabilities
> Use the release-pm agent to review changelog entries
> Use the release-docs agent to verify documentation is updated
```

## Available Agents

| Agent | Description |
|-------|-------------|
| `release-coordinator` | Orchestrates the complete release workflow |
| `release-pm` | Product manager perspective for release planning |
| `release-qa` | Quality assurance and test verification |
| `release-security` | Security review and vulnerability checking |
| `release-docs` | Documentation completeness review |

## Agent Details

### release-coordinator

Main orchestrator that:

- Determines semantic version from conventional commits
- Generates changelog via schangelog
- Runs validation checks (build, test, lint)
- Verifies CI status before tagging
- Creates and pushes release tags

### release-pm

Reviews releases from product perspective:

- Validates changelog entries are user-friendly
- Checks feature completeness against roadmap
- Reviews breaking changes communication

### release-qa

Ensures quality standards:

- Verifies test coverage
- Checks for test failures
- Reviews error handling

### release-security

Security-focused review:

- Checks for hardcoded secrets
- Reviews dependency vulnerabilities
- Validates input sanitization

### release-docs

Documentation completeness:

- Verifies README is updated
- Checks API documentation
- Reviews migration guides for breaking changes

## Dependencies

The agents expect these CLI tools to be installed:

- `releaseagent` - Release automation CLI
- `schangelog` - Structured changelog generator
- `sroadmap` - Roadmap management tool
- `golangci-lint` - Go linter (for Go projects)

## Steering Files

The `steering/` directory contains context files that enhance agent behavior:

| File | Purpose |
|------|---------|
| `release-workflow.md` | Complete release process documentation |
| `version-analysis.md` | Semantic versioning guidelines |
| `commit-classification.md` | Conventional commits classification |

Copy these to `.kiro/steering/` in your project for automatic context loading.
