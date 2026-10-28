# ldvdt — trainer for the `ldvd` keyboard layout

A small local web server that teaches and drills the custom **`ldvd`** keyboard
layout (a Latin-American Dvorak variant adapted for programming) through
cumulative difficulty levels.

The binary serves a server-rendered SPA, stores your progress in SQLite, and
opens your default browser.

The page is a single Bootstrap container with three stacked rows:

```
┌──────────────────────────────────────────────────────────────┐
│ Level · Language · Word count · [ Empezar ]      [ Theme ]   │
├──────────────────────────────────────────────────────────────┤
│                  isla  otro  sol  ...                        │
│            ISO keyboard, drawn as an inline SVG              │
├──────────────────────────────────────────────────────────────┤
│ Progress: stats + chart                                      │
└──────────────────────────────────────────────────────────────┘
```

## Requirements

- Go 1.27+ (no CGO needed: the SQLite driver is pure Go)
- `templ` (only when regenerating templates): `go install github.com/a-h/templ/cmd/templ@v0.3.1070`
- Node.js (only for the JavaScript tests)
- Dart Sass (only when rebuilding the vendored Bootstrap theme, see below)

## Run

```bash
go run ./cmd/ldvdt
```

The server listens on <http://127.0.0.1:10987> and opens the default browser.

## Build

```bash
go build -o ldvdt ./cmd/ldvdt
```

Cross-compilation works out of the box because nothing requires CGO:

```bash
GOOS=windows GOARCH=amd64 go build -o ldvdt.exe ./cmd/ldvdt
GOOS=linux   GOARCH=arm64 go build -o ldvdt     ./cmd/ldvdt
```

## Tests

```bash
go test ./...                    # Go: domain, corpus, storage, handlers, UI, browser
node test/app_test.js            # dead-key composition in the typing engine
node test/chart_test.js          # progress chart rendering
node test/keyboard_test.js       # SVG keyboard highlight and restore
node test/new_practice_test.js   # "Nueva práctica" restart button
node test/theme_test.js          # light/dark theme toggle
```

The JS suites are plain Node scripts with no dependencies; CI runs every
`test/*.js`.

## Code generation

The keyboard layout comes from `docs/layout.yaml` **at build time** and is then
embedded in the binary; the YAML is never read at runtime.

```bash
go generate ./internal/domain/   # layout_gen.go + the embedded SVG copy
templ generate                   # *.templ -> *_templ.go
```

| Source | Generated | What it is |
| ------ | --------- | ---------- |
| `docs/layout.yaml` | `internal/domain/layout_gen.go` | the fixed 48-key layout, including the `id` that names each key in the SVG |
| `docs/keyboard.svg` | `internal/web/ui/keyboard.svg` | the keyboard design, embedded inline in the page (XML prolog stripped) |

Commit the generated files: CI runs both commands and fails on a dirty diff.
Re-run `go generate` after any change to `docs/layout.yaml` or
`docs/keyboard.svg`.

## Bootstrap theme

Bootstrap is vendored **compiled from Sass**, not patched with overrides:
Bootstrap 5.3 hardcodes the palette inside its components (`.btn-primary`,
the form focus ring), so overriding `--bs-primary` alone would leave them on the
defaults.

```bash
npm pack bootstrap@5.3.3 && tar -xzf bootstrap-5.3.3.tgz   # provides package/scss
npx sass --no-source-map --no-charset --style=compressed --quiet-deps \
  --load-path=package/scss \
  docs/bootstrap-theme.scss internal/web/static/vendor/bootstrap.min.css
```

`docs/bootstrap-theme.scss` holds the palette (`primary: #26a269`, the rest the
Bootstrap defaults) and is the only file to edit. Only the CSS is vendored: the
theme switch uses `data-bs-theme`, which is pure CSS, and no JavaScript
component (dropdown, modal, tooltip) is used.

## Architecture

Hexagonal, wired through [hefesto](https://github.com/wfrs-dev/hefesto) (a small
Echo v5 wrapper):

```
cmd/ldvdt                 main: config, storage, corpus, server, browser
internal/domain           Keyboard, Key, Level, Practice, Stats  (pure)
internal/corpus           embedded word lists + level-aware word selection
internal/port             Storage port (interface)
internal/infra/storage    SQLite adapter (modernc.org/sqlite)
internal/infra/browser    opens the default browser (Linux/Windows/macOS)
internal/trainer          HTTP handlers + use cases + routes
internal/web/ui           templ components (page, keyboard SVG, progress)
internal/web/static       embedded CSS/JS + vendored Bootstrap, htmx and Chart.js
internal/tools/genlayout  layout.yaml -> Go literal + SVG copy generator
resources                 embedded practice corpora (eng.dat, spa.dat)
```

Request flow: `handler → Service (use cases) → port.Storage → sqlite`.

### HTTP endpoints

| Method | Path        | Purpose                                        |
| ------ | ----------- | ---------------------------------------------- |
| GET    | `/`         | Full SPA shell (progress, words, keyboard)     |
| GET    | `/practice` | Word fragment for `level`, `lang`, `count`     |
| POST   | `/practice` | Persist a finished practice, return stats      |
| GET    | `/progress` | Progress panel fragment (stats + chart series) |
| GET    | `/status/health` | Version, OS, uptime (from hefesto)        |
| GET    | `/static/*` | Embedded assets                                |

## Levels

Levels are cumulative: level *N* allows every character introduced up to *N*.

| Level | Adds |
| ----- | ---- |
| 1 | Home row, lowercase: `a o e i u d r t n s` |
| 2 | Home row, uppercase: `A O E I U D R T N S` |
| 3 | Upper row lowercase: `ñ , . p y f g c h l` + accents `á é í ó ú` |
| 4 | Upper row shifted: `Ñ ; : P Y F G C H L` + accents `Á É Í Ó Ú` |
| 5 | Bottom row unshifted: `* - q j k x b m w v z` |
| 6 | Bottom row shifted: `% _ Q J K X B M W V Z` |
| 7 | Numeric row shifted: `~ 1 2 3 4 5 6 7 8 9 0 ? =` |
| 8 | Numeric row unshifted: `$ & [ { ( < > ) } ] /` plus the backtick and `+` |

Words come from `resources/spa.dat` and `resources/eng.dat`. When the selected
level has too few eligible words, the remaining practice text is generated at
random (5–10 characters) from the level's allowed characters. Numeric lines in
the corpora are filtered out.

## Layout semantics

Every key in `docs/layout.yaml` has two characters:

- `bottom` — produced by a plain key press
- `top` — produced by holding **Shift**

Accented characters are typed with the dead key `´` followed by the vowel
(`´` + `a` → `á`, `´` + `A` → `Á`). The on-screen keyboard highlights **both**
keys while an accented character is pending.

## Metrics

| Metric | Definition |
| ------ | ---------- |
| WPM | characters ÷ 5 per minute |
| Accuracy | correct keystrokes ÷ total keystrokes (0..1) |
| Consistency | standard deviation of the interval between keystrokes, in ms |
| Duration | wall-clock time of the practice, in ms |

Per practice the app stores language, level, WPM, accuracy, consistency,
duration and word count; the progress panel aggregates total practices, total
time, best/average WPM, average accuracy and average consistency.

## Data

SQLite database location (`os.UserConfigDir()`):

| OS | Path |
| -- | ---- |
| Linux | `~/.config/ldvdt/ldvdt.db` |
| Windows | `%AppData%\ldvdt\ldvdt.db` |
| macOS | `~/Library/Application Support/ldvdt/ldvdt.db` |

The database is created on first run.

## Notes

- The server binds to `:10987` on **all interfaces** (hefesto's `Start` hardcodes
  `":" + port`). The app has no authentication, so treat it as a local tool.
- Typing is captured with `keydown` on the page; the browser window must have
  focus while practising.
- When a dead key is used, browsers may deliver either the composed character
  (`á`) or `Dead` followed by the base letter (`a`). Both are handled.
