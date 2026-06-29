# Dino Claimable Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an offline `arksave dino-claimable` command that reports owned/tamed dinos whose claim timer has expired.

**Architecture:** Mirror `structure-demolishable`: add an `arkapi` report API, use selected-property reads for large-save behavior, expose stable JSON/table CLI output, preserve raw timing fields for auditability, and document that this is an offline computed report rather than live-server truth. Use `GameTime` as now, `LastInAllyRangeSerialized` as the primary claim reset timestamp, and `LastInAllyRangeTimeSerialized`/`TamedTimeStamp` as visible fallbacks when the primary ally-range field is absent.

**Tech Stack:** Go, existing `arksave`/`arkobject` parsers, `cmd/arksave` CLI, synthetic SQLite save fixtures, `go test`, `make verify`.

---

### Task 1: Dino Model Timing Fields

**Files:**
- Modify: `arkobject/dino.go`
- Modify: `internal/testfixtures/fixtures.go`
- Test: `arkobject/dino_test.go`

- [x] Add `TamedTimeStamp`, `LastInAllyRangeSerialized`, and `LastInAllyRangeTimeSerialized` to `arkobject.Dino`.
- [x] Extend `DinoFromObject` to read both fields.
- [x] Extend `DinoGameObjectOptions` and `DinoGameObjectBytes` to emit `LastInAllyRangeSerialized` and `LastInAllyRangeTimeSerialized`.
- [x] Add/adjust tests proving both fields parse from object properties.

### Task 2: Claimable Report API

**Files:**
- Create: `arkapi/dino_claim.go`
- Create: `arkapi/dino_claim_test.go`

- [x] Add options for map name, claim period, claim multiplier, and optional `GameUserSettings.ini`.
- [x] Parse `PvEDinoDecayPeriodMultiplier` from settings when supplied; default multiplier to `1`.
- [x] Default claim period to 8 days and allow `--claim-period` seconds to override it.
- [x] Filter to dinos with ownership signals, excluding dead and cryopodded dinos.
- [x] Compute from `LastInAllyRangeSerialized`, falling back to `LastInAllyRangeTimeSerialized` and then `TamedTimeStamp`.
- [x] Sort by owner/tribe, then location, then dino short name.
- [x] Include raw timing fields plus `claim_reference_time` and `claim_reference_source`.

### Task 3: CLI And Redaction

**Files:**
- Modify: `cmd/arksave/main.go`
- Modify: `cmd/arksave/main_test.go`

- [x] Add `dino-claimable` usage and dispatch.
- [x] Support `--game-user-settings`, `--claim-multiplier`, `--claim-period`, `--map`, and `--json`.
- [x] Print table output sorted by owner/location.
- [x] Emit stable ordered JSON.
- [x] Redact owner IDs, dino UUIDs, dino IDs, and paths under `--redact`.
- [x] Add `--oldest N` diagnostics for saves with owned dinos but no claimable rows.
- [x] Return clear errors for missing save path, bad multiplier, bad period, and unreadable INI.

### Task 4: Docs And Verification

**Files:**
- Modify: `README.md`
- Modify: `docs/project-task-ledger.md`
- Modify: `docs/task-inventory.md`

- [x] Document `dino-claimable`, supported options, and offline-computed limitations.
- [x] Update task ledgers so progress is visible from the repo.
- [x] Run `go test ./arkobject ./arkapi ./cmd/arksave`.
- [x] Run `make verify`.
- [x] Commit and push to `main`.
