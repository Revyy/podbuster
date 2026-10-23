# GitHub Copilot Instructions

See [AGENTS.md](../AGENTS.md) for full project context, coding conventions, and package responsibilities.

## Quick Reference

- `game/` must never import `cluster/` or `bridge/`
- Cross-layer comms: typed channels in `bridge/`
- Run: `go run . --mock`
- Test: `go test ./...`
- Validate before a PR: `make check` (same as CI)
