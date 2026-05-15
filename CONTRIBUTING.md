# Contributing to tango-go

Thanks for taking the time to contribute! `tango-go` is the official Go SDK for the [Tango API](https://makegov.com). It's deliberately small, idiomatic, and aligned with its sibling SDKs ([`tango-node`](https://github.com/makegov/tango-node) and [`tango-python`](https://github.com/makegov/tango-python)).

## Ground rules

- The public API is stable. Breaking changes ship in major releases, not patches.
- New endpoints should land alongside their typed input/output structs and at least one example.
- godoc on every exported identifier.
- No new external runtime dependencies without a discussion in an issue first.

## Local development

You'll need Go 1.23+ and [`golangci-lint`](https://golangci-lint.run/) on your PATH.

```bash
git clone https://github.com/makegov/tango-go.git
cd tango-go
make test      # unit tests
make lint      # golangci-lint
make ci        # vet + race tests + coverage + lint (matches CI)
```

See `make help` for the full list of targets.

### Coverage

```bash
make cover       # coverage.out + per-function summary
make cover-html  # coverage.html (open it in a browser)
```

### Integration tests

Integration tests live under `tests/integration/` behind the `integration` build tag and hit the live Tango API. They require `TANGO_API_KEY` in your environment.

```bash
TANGO_API_KEY=... make integration
```

## Branch and commit conventions

- Branch from `main`. Branch names: `<type>/<short-description>` (e.g. `feat/idv-subresources`, `fix/retry-jitter`).
- Use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages. Common types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`.
- Reference the relevant issue in the body (e.g. `Closes #42`).

## Pull request checklist

- [ ] `make ci` is green locally.
- [ ] New exported identifiers have godoc.
- [ ] `CHANGELOG.md` has an entry under `## [Unreleased]` in the appropriate section (`Added`, `Changed`, `Fixed`, `Removed`).
- [ ] If the change is user-facing, `docs/` and `README.md` are updated.
- [ ] No new external runtime dependencies (or you've gotten a thumbs-up in an issue first).

## Releasing

Releases are tag-driven. Pushing a `vX.Y.Z` tag triggers `.github/workflows/release.yml`, which runs the test suite, creates a GitHub Release with auto-generated notes, and pings `proxy.golang.org` so pkg.go.dev indexes the new tag.

## Reporting issues

Bug? Feature request? Use the templates in `.github/ISSUE_TEMPLATE/`. For security issues, see [SECURITY.md](SECURITY.md) — do **not** open a public issue.
