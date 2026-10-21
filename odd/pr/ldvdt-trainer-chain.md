# Delivery plan — `feat/ldvdt-trainer` (feature-branch chain)

Strategy: **`chain_strategy=feature-branch-chain`**, selected by the project owner.

## Status

**Nothing is pushed and no PR exists.** This repository has no `git remote`
configured, so the target host, `owner/repo` and default branch are unknown, and
no issue can be linked. Every remote operation below needs its own explicit
authorization.

## Review budget

`master..HEAD` = 90 files, **+10,751 / -0** lines:

| Category | Files | Lines |
| --- | --- | --- |
| Authored source and docs | 46 | 4,518 |
| Practice corpora (`resources/*.dat`, data) | 2 | 6,233 |
| Binary theme assets (`resources/font/**`) | 42 | — |

Review budget (the number that matters for slicing) is the **4,518 authored
lines**, not the 10,751 total.

## Slices

Branches are cumulative; each child includes its ancestors.

| # | Branch | Commits | Authored lines | Depends on |
| --- | --- | --- | --- | --- |
| 1 | `feat/ldvdt-foundation` | `a2cc5a3..408c71f` | ~1,586 | — |
| 2 | `feat/ldvdt-practice-slice` | `aea4dba` | ~1,581 | PR 1 |
| 3 | `feat/ldvdt-keyboard-accent` | `f1e2f93..cc62f01` | ~733 | PR 2 |
| 4 | `feat/ldvdt-charts-docs` | `cf44dbb..ce8598c` | ~771 | PR 3 |

No cohesive split brings every slice under the advisory 400-line budget without
deleting tests or docs, or splitting a work unit from the tests that verify it.
The four slices above are the single honest slicing pass: one deliverable per
slice, tests and docs kept with the unit they verify.

### PR 1 — Foundation

- **Scope:** module scaffold, `docs/layout.yaml` home-row correction, domain
  keyboard/level model and generator, corpus loader, storage port and SQLite
  adapter, configuration, server bootstrap, embedded asset wiring.
- **End state:** the binary starts and serves `/status/health`; no practice UI.
- **Out of scope:** HTTP trainer routes, templ UI, typing engine.

```
feature tracker: feat/ldvdt-trainer (draft)
└── PR 1  feat/ldvdt-foundation  📍
```

### PR 2 — Practice slice

- **Scope:** templ SPA shell and ISO keyboard, stylesheet, client typing engine,
  trainer handlers and use cases, end-to-end level 1 test.
- **End state:** a full practice works from the browser for level 1.
- **Out of scope:** accents, charts.

```
feature tracker: feat/ldvdt-trainer (draft)
└── PR 1  feat/ldvdt-foundation
    └── PR 2  feat/ldvdt-practice-slice  📍
```

### PR 3 — Keyboard polish and accented characters

- **Scope:** word spacing and wrapping, spacebar highlight, decorative ISO
  modifier keys, uniform row widths with the real ISO Enter and super keys,
  dead-key accent support and the browser dead-key composition fix.
- **End state:** levels 3 and 4 are playable, including `á é í ó ú` and `Á É Í Ó Ú`.
- **Out of scope:** charts, README.

```
feature tracker: feat/ldvdt-trainer (draft)
└── PR 1  feat/ldvdt-foundation
    └── PR 2  feat/ldvdt-practice-slice
        └── PR 3  feat/ldvdt-keyboard-accent  📍
```

### PR 4 — Charts, docs and review fixes

- **Scope:** Chart.js progress charts and series query, README, cross-platform
  browser launcher with its test, the approved native review correction, and the
  advisory follow-ups.
- **End state:** the feature is complete.
- **Review note:** `resources/font/**` (42 binaries) and `resources/colors/**`
  travel in commit `845806e`; consider moving them to their own commit or PR
  before submission if the owner prefers a cleaner slice.

```
feature tracker: feat/ldvdt-trainer (draft)
└── PR 1  feat/ldvdt-foundation
    └── PR 2  feat/ldvdt-practice-slice
        └── PR 3  feat/ldvdt-keyboard-accent
            └── PR 4  feat/ldvdt-charts-docs  📍
```

## Preparation commands (local, no remote writes)

The current branch holds every commit, so free the tracker name first:

```bash
git branch -m feat/ldvdt-trainer feat/ldvdt-candidate   # keep all work
git branch feat/ldvdt-trainer master                    # tracker, empty
git branch feat/ldvdt-foundation      408c71f
git branch feat/ldvdt-practice-slice  aea4dba
git branch feat/ldvdt-keyboard-accent cc62f01
git branch feat/ldvdt-charts-docs     ce8598c
```

## Delivery commands (require explicit authorization, one per action)

```bash
# 1. Authorized target
git remote add origin <host>:<owner>/ldvdt.git

# 2. Tracker branch (no PR content; draft/no-merge tracker PR targets the default branch)
git push -u origin feat/ldvdt-trainer

# 3. Children, in order; each PR targets the branch above it
git push -u origin feat/ldvdt-foundation
git push -u origin feat/ldvdt-practice-slice
git push -u origin feat/ldvdt-keyboard-accent
git push -u origin feat/ldvdt-charts-docs

# 4. PRs (base is the branch one row up; PR 1 targets the tracker)
gh pr create --draft --repo <owner>/ldvdt --base feat/ldvdt-trainer \
  --head feat/ldvdt-foundation --title "feat: ldvdt foundation" \
  --body-file odd/pr/ldvdt-trainer-pr.md
# ... repeat for each child, changing --base/--head
```

## Blockers before any of this

1. **No `git remote`** — the target host and `owner/repo` are undefined.
2. **No approved issue** — the target policy requires `status:approved`; the body
   carries a `Refs #<issue>` placeholder.
3. **No `type:*` label evidence** — labels are read from the target repository.
4. **No default branch verified** — the tracker base must come from target
   metadata, not from an assumed `main`.
5. **`845806e` mixes owner assets with a docs commit** — decide whether to move
   the 42 binaries out first (history rewrite, needs authorization).
