     ╔══════════════════════════════════════════════════════════════════════════════╗
     ║                       RELEASE VALIDATION STARTING                            ║
     ╚══════════════════════════════════════════════════════════════════════════════╝

     ▶ Running PM validation...
     ▶ Running QA validation...
     ▶ Running Documentation validation...
     ▶ Running Release Management validation...
     ▶ Running Security validation...
     ╔══════════════════════════════════════════════════════════════════════════════╗
     ║                              TEAM STATUS REPORT                              ║
     ╠══════════════════════════════════════════════════════════════════════════════╣
     ║ Project: github.com/plexusone/agent-team-release                           ║
     ║ Target:  v0.6.0                                                              ║
     ╠══════════════════════════════════════════════════════════════════════════════╣
     ║ PHASE 1: REVIEW                                                              ║
     ╠══════════════════════════════════════════════════════════════════════════════╣
     ║ pm-validation (pm)                                                           ║
     ║   version-recommendation   🟢 GO    v0.6.0 appropriate for minor (featu...   ║
     ║   release-scope            🟢 GO    11 changes documented                    ║
     ║   changelog-quality        🟢 GO    2 highlights present                     ║
     ║   breaking-changes         🟢 GO    3 breaking changes documented            ║
     ║   roadmap-alignment        🟢 GO    No roadmap items tagged for v0.6.0       ║
     ║   deprecation-notices      🟢 GO    No deprecations                          ║
     ╠══════════════════════════════════════════════════════════════════════════════╣
     ║ docs-validation (documentation)                                              ║
     ║   readme.md                🟢 GO                                             ║
     ║   prd.md                   🟢 GO                                             ║
     ║   trd.md                   🟢 GO                                             ║
     ║   mkdocs-site              🟢 GO                                             ║
     ║   release-notes            🟢 GO    Found: docs/releases/v0.6.0.md           ║
     ║   changelog.md             🟢 GO                                             ║
     ╠══════════════════════════════════════════════════════════════════════════════╣
     ║ qa-validation (qa)                                                           ║
     ║   go:no-local-replace      🟢 GO                                             ║
     ║   go:mod-tidy              🟢 GO                                             ║
     ║   go:build                 🟢 GO                                             ║
     ║   go:format                🔴 NO-GO cmd/atrelease/check.go                   ║
     ║   go:lint                  🔴 NO-GO pkg/checks/releasekit.go:151:6: fun...   ║
     ║   go:test                  🟢 GO                                             ║
     ║   go:vulncheck             🟢 GO                                             ║
     ║   go:error-handling        🔴 NO-GO pkg/git/git.go:224: potential error...   ║
     ║   go:untracked-refs        🟡 WARN  cmd/atrelease/check.go may referenc...   ║
     ╠══════════════════════════════════════════════════════════════════════════════╣
     ║ security-validation (security)                                               ║
     ║   license-file             🟢 GO    LICENSE                                  ║
     ║   vulnerability-scan       🟢 GO    No vulnerabilities found                 ║
     ║   dependency-audit         🟢 GO                                             ║
     ║   no-hardcoded-secrets     🟢 GO                                             ║
     ╠══════════════════════════════════════════════════════════════════════════════╣
     ║ release-validation (release)                                                 ║
     ║   version-available        🟢 GO    Tag v0.6.0 is available                  ║
     ║   git-working-directory    🟡 WARN  58 uncommitted changes                   ║
     ║   git-remote               🟢 GO    https://github.com/agentplexus/agen...   ║
     ║   changelog.json           🟢 GO                                             ║
     ║   ci-configuration         🟢 GO    Found: .github/workflows                 ║
     ╠══════════════════════════════════════════════════════════════════════════════╣
     ║                         🛑 TEAM: NO-GO for v0.6.0 🛑                         ║
     ╚══════════════════════════════════════════════════════════════════════════════╝