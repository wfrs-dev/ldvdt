<!-- Prepared PR candidate for feat/ldvdt-trainer.
     Not deliverable yet: no git remote is configured and no approved issue exists.
     Pending: choose the delivery strategy (single PR vs chained PRs). -->

## Linked issue

<!-- BLOCKER: the target policy requires an issue carrying status:approved.
     Replace the placeholder and pick the reference form with the human:
     Closes/Fixes/Resolves for a closing reference, Refs for non-closing. -->
Refs #<issue>

## PR type

- [x] New feature

<!-- Add exactly one label: type:feature -->

## Summary

- Adds `ldvdt`, a local web trainer for the custom `ldvd` keyboard layout: a Go
  binary (hefesto/Echo v5, hexagonal architecture) that serves a server-rendered
  SPA (templ + htmx) and opens the default browser on port 10987.
- Implements eight cumulative levels over the ISO layout generated at build time
  from `docs/layout.yaml`, level-aware word selection from the embedded
  Spanish/English corpora with a random 5–10 character fallback, and a client
  typing engine that measures WPM, accuracy and keystroke-rhythm consistency.
- Persists progress in SQLite under the user's config directory and renders the
  WPM/accuracy history with Chart.js; every asset is embedded in the binary.

## Changes

| Area | Change |
| --- | --- |
| `cmd/ldvdt` | Server bootstrap, configuration, browser launch |
| `internal/domain` | Keyboard, key, finger and level model; generated layout literal |
| `internal/corpus` | Embedded corpora loader and level-aware word selection |
| `internal/port`, `internal/infra/storage` | Storage port and SQLite adapter |
| `internal/trainer` | HTTP handlers, use cases and routes |
| `internal/web/ui` | templ components: page, ISO keyboard, progress panel |
| `internal/web/static` | Stylesheet, typing engine, vendored htmx and Chart.js |
| `internal/infra/browser` | Cross-platform default-browser launcher |
| `internal/tools/genlayout` | `docs/layout.yaml` to Go literal generator |
| `resources` | Practice corpora, theme colours and fonts |
| `docs/layout.yaml` | Home row corrected to `bottom` = plain, `top` = shifted |
| `README.md` | Requirements, run/build, endpoints, levels, metrics, data paths |

## Test plan

| Command | Outcome |
| --- | --- |
| `go build ./...` | pass |
| `go vet ./...` | clean |
| `go test ./...` | 6 packages pass |
| `node test/app_test.js` | dead-key composition assertions pass |
| `node test/chart_test.js` | chart rendering assertions pass |
| `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./cmd/ldvdt` | pass |
| `GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build ./cmd/ldvdt` | pass |
| Binary smoke test | `/` renders the SPA, `/practice` returns level words, `POST /practice` persists and returns updated stats |
| Native review (RDD on) | lineage `review-658f904231a38014` approved after correcting one critical finding |

Shellcheck: not applicable — no shell scripts changed.

## Review notes

- `resources/font/**` (42 binary files) and `resources/colors/**` are theme
  assets coming from the project owner. They were captured in commit `845806e`
  together with the review log and are **not used by the trainer yet**; they
  can be moved to their own commit or PR before submission if preferred.
- `resources/*.dat` are the practice corpora (6,233 lines of data), not source.
- Accented characters rely on the browser delivering either the composed
  character or `Dead` plus the base letter; both paths are handled and pinned by
  `test/app_test.js`.

## Contributor checklist

- [x] Conventional commit messages
- [x] No `Co-Authored-By` trailers
- [x] Tests added or updated with the behaviour they cover
- [x] Documentation updated (`README.md`)
- [ ] Linked issue exists and carries `status:approved` — **pending**
- [ ] Exactly one `type:*` label added — **pending** (needs the target repository)
