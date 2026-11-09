---
title: "Development"
description: "Build, lint and test ctxerrors with its Make targets, the 90 percent coverage gate, the vendored golangci-lint tool, and how CI and releases run."
nav_order: 7
---

# Development

Hacking on ctxerrors itself. You need Go 1.26 or newer and Make. Everything else, golangci-lint included, comes from the module's own `tool` dependencies in [`go.mod`](https://github.com/psyb0t/ctxerrors/blob/main/go.mod) and is vendored in [`vendor/`](https://github.com/psyb0t/ctxerrors/tree/main/vendor), so there is nothing to install globally.

## Make targets

Run `make help` for the list. The [`Makefile`](https://github.com/psyb0t/ctxerrors/blob/main/Makefile) has:

| Target | What it does |
| --- | --- |
| `make all` | Runs `dep`, `lint` and `test`, in that order. |
| `make dep` | `go mod tidy` and `go mod vendor`. |
| `make lint` | Fails if `go fix -diff ./...` has suggestions, then runs `go tool golangci-lint run ./...` with the rules in [`.golangci.yml`](https://github.com/psyb0t/ctxerrors/blob/main/.golangci.yml). |
| `make lint-fix` | Applies `go fix ./...` and runs golangci-lint with `--fix`. |
| `make test` | `go test -race ./...`. |
| `make test-coverage` | Runs the tests with the race detector and a coverage profile, writes the total to `coverage-percent.txt`, and fails below 90 percent. |

[`pre-commit.sh`](https://github.com/psyb0t/ctxerrors/blob/main/pre-commit.sh) runs `make lint && make test-coverage`, the same pair CI runs.

## Layout

| Path | Holds |
| --- | --- |
| [`ctxerrors.go`](https://github.com/psyb0t/ctxerrors/blob/main/ctxerrors.go) | `CTXError`, `New`, `Wrap`, `Wrapf`, `Join` and the caller lookup. |
| [`errormap.go`](https://github.com/psyb0t/ctxerrors/blob/main/errormap.go) | The process-wide error map and its three functions. |
| [`commerr/`](https://github.com/psyb0t/ctxerrors/tree/main/commerr) | The common sentinel errors. |
| [`docs/`](https://github.com/psyb0t/ctxerrors/tree/main/docs) | This site, built by GitHub Pages from the `main` branch. |

The tests are internal (`package ctxerrors`) and live next to the code in `*_internal_test.go`.

## CI

[`.github/workflows/pipeline.yml`](https://github.com/psyb0t/ctxerrors/blob/main/.github/workflows/pipeline.yml) runs on pushes to any branch, on `v*` tags, once a week on a schedule, and by hand.

- On every branch push, the schedule and manual runs, the shared Go workflow runs `make lint`, `make test-coverage` and govulncheck.
- When that run is on `main`, the badges job then publishes the coverage, version, license and imported-by badges to the `badges` branch.
- On a `v*` tag, the pipeline does not rerun the checks. It waits for the `main` run of the same commit to pass, cuts the GitHub release with the tag message as its notes, and refreshes the badges with the coverage from that `main` run. The `main` run keeps its coverage artifact for one day, so a tag pushed later than that still gets its release, but the badges job fails.

The weekly run exists for the imported-by badge, which pkg.go.dev updates on its own schedule.

## Releases

Versions follow [semver](https://semver.org) pre-1.0 conventions, as the [CHANGELOG](https://github.com/psyb0t/ctxerrors/blob/main/CHANGELOG.md) says at the top: a minor bump may break the API and calls it out, a patch bump is docs, build or fixes only. The Git tag is the version; there is no version file. A release is a commit on `main` that has passed CI, then an annotated `vX.Y.Z` tag on it whose message is that version's CHANGELOG section.
