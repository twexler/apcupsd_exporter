# Code Context

## Files Retrieved
1. `go.mod` (lines 1-200) - module path & Go version
2. `.github/workflows/linux-test.yml` (lines 1-200) - CI test workflow
3. `.github/workflows/static-analysis.yml` (lines 1-200) - staticcheck/vet workflow
4. `.github/workflows/goreleaser.yml` (lines 1-200) - release workflow using Go 1.15
5. `.goreleaser.yml` (lines 1-200) - goreleaser config
6. `README.md` (lines 1-200) - user/docs and badges
7. `apcupsdexporter.go` (lines 1-200) - Exporter entry/collector orchestration
8. `upscollector.go` (lines 1-400) - main collector implementation
9. `apcupsdexporter_test.go` (lines 1-200) - test harness
10. `upscollector_test.go` (lines 1-400) - collector tests
11. `cmd/` (listing) - main binary entrypoint (cmd/apcupsd_exporter/main.go)

## Key Code
- Exporter type and withCollectors: apcupsdexporter/apcupsdexporter.go (lines 1-200)
- UPSCollector and metrics: apcupsdexporter/upscollector.go (lines 1-400)
- Tests exercising collector registration and metric output: upscollector_test.go (lines 1-400)

## Repo Summary (brief)
- Language: Go
- Module path: github.com/mdlayher/apcupsd_exporter (go.mod)
- Main packages: package apcupsdexporter and cmd/apcupsd_exporter
- Notable files: go.mod, .github workflows, .goreleaser.yml, README.md, apcupsdexporter.go, upscollector.go, tests

## Current test status
- Local: go test ./... -> all tests pass (ok github.com/mdlayher/apcupsd_exporter)

## Immediate red flags
- Mixed Go versions in CI: go.mod says 1.17, some workflows use 1.15 (goreleaser/publish_build).
- Dependency pins include pseudo-version for github.com/mdlayher/apcupsd (2023) — check for newer tags.
- .goreleaser.yml uses GO111MODULE env; workflows use old Go (1.15) — modernization needed.
- No failing tests locally; CI config exists but versions inconsistent.

## Prioritized next actions (one-line each)
- go-dep-updater: bump module to latest stable Go (1.20/1.21) in go.mod and update deps (check apcupsd tag).
- ci-setup: unify GitHub Actions to a single modern Go matrix (e.g., 1.20,1.21) and remove 1.15 jobs.
- linter-fixer: run golangci-lint or go vet/staticcheck locally, fix reported issues (staticcheck already in CI).
- static-checks: run staticcheck@latest and go vet, add results to CI gating and fix blockers.
- docs-and-badges: update README badges to point to current workflows and GoDoc; add Go version badge.
- release-automation: update goreleaser workflow to use the repo Go version and modern action versions; remove GO111MODULE env.

## Files to open for manual review
- go.mod — verify go version and dependency versions
- .github/workflows/* — normalize Go versions and gating rules
- .goreleaser.yml & .github/workflows/goreleaser.yml — align Go version and build steps
- cmd/apcupsd_exporter/main.go — flags, defaults, and build entrypoint
- apcupsdexporter.go & upscollector.go — collector lifecycle and error handling
- upscollector_test.go & apcupsdexporter_test.go — test coverage and flaky tests

Start here: open go.mod then .github/workflows/linux-test.yml for CI version alignment.
