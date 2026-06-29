# Focused Oracle Probes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add narrow, throwaway-style Python oracle extractors that unblock specific Go parity questions without expanding or improving the upstream Python codebase.

**Architecture:** Keep focused probes in this repository under `scripts/`, reuse the pinned upstream checkout in `.oracle/upstream/src`, and write all detailed values under ignored `.oracle/output`. Commit only harness code, tests, and sanitized status documentation.

**Tech Stack:** Python `unittest`, pinned upstream `ark-save-parser`, Go `make verify`, private `.oracle` workspace.

---

### Task 1: Probe Harness

**Files:**
- Create: `scripts/oracle_probe.py`
- Create: `scripts/oracle_probe_test.py`
- Modify: `Makefile`
- Modify: `docs/development.md`
- Modify: `docs/task-inventory.md`

- [x] **Step 1: Write failing summary tests**

Create `scripts/oracle_probe_test.py` with tests for `ProbeResult`, `summary_rows`, and `default_output_path`.

- [x] **Step 2: Run test to verify it fails**

Run:

```sh
PYTHONPYCACHEPREFIX="$PWD/.cache/pycache" python3 -m unittest scripts/oracle_probe_test.py
```

Expected: fail with `ModuleNotFoundError: No module named 'oracle_probe'`.

- [x] **Step 3: Add minimal probe module**

Create `scripts/oracle_probe.py` with:

```python
@dataclass(frozen=True)
class ProbeResult:
    name: str
    status: str
    detail: str
    output: str
```

The CLI must require `ARK_ORACLE_SAVE`, accept `--probe equipment_rank`, `--probe dino_cryopod_location`, or `--probe all`, write detailed private JSON to `.oracle/output/oracle-probe-<probe>.json`, and print only JSON summary rows with probe names, status, and detail.

- [x] **Step 4: Run unit tests to verify they pass**

Run:

```sh
PYTHONPYCACHEPREFIX="$PWD/.cache/pycache" python3 -m unittest scripts/oracle_probe_test.py
```

Expected: pass.

- [x] **Step 5: Wire Makefile and docs**

Add `make oracle-probe` and document that these probes are permitted as a narrow scope extension for concrete parity blockers only.

- [x] **Step 6: Run full verification**

Run:

```sh
make verify
```

Expected: pass.

- [x] **Step 7: Commit and push**

Commit:

```sh
git add Makefile docs/development.md docs/task-inventory.md docs/superpowers/plans/2026-06-29-focused-oracle-probes.md scripts/oracle_probe.py scripts/oracle_probe_test.py
git commit -m "Add focused oracle probe harness"
git push origin main
```

### Task 2: First Probe-Backed Parity Slice

**Files:**
- Modify: `cmd/arksave/main.go`
- Modify: `cmd/arksave/main_test.go`
- Modify: `docs/task-inventory.md`
- Modify: `docs/phase-2-blocker-matrix.md`
- Modify: `docs/project-task-ledger.md`
- Modify: `docs/phase-2-transpilation.md`
- Modify: `docs/oracle-comparison-summary.md`

- [x] **Step 1: Run one focused private probe**

Run:

```sh
ARK_ORACLE_SAVE=/absolute/path/to/private/save.ark make oracle-probe ORACLE_PROBE_ARGS="--probe equipment_rank"
```

Expected: committed stdout contains only status metadata; detailed private values stay in `.oracle/output`.

- [x] **Step 2: Choose one concrete mismatch or unblocker**

The first probe selected the `equipment-rank` candidate filter. The Go command had corrected the upstream example's mixed-case `CLoth` ignore token into `Cloth`, which incorrectly filtered normal cloth armor. Do not commit private values from `.oracle/output`.

- [x] **Step 3: Add a Go failing test using synthetic or sanitized inputs**

Extend `TestEquipmentRankCommandPrintsRankStats` so a synthetic cloth armor blueprint remains ranked. Before the Go fix, the test fails with `Ranked: 2` and `Classes: 2` instead of the expected `Ranked: 3` and `Classes: 3`.

- [x] **Step 4: Implement minimal Go parity fix**

Change `ignoredEquipmentNameParts` in `cmd/arksave/main.go` from `Cloth` to the upstream example's exact `CLoth` token.

- [x] **Step 5: Verify and document**

Run focused Go tests, `make verify`, update status docs, commit, push, and watch CI.

## Self-Review

- Spec coverage: This plan implements the newly allowed narrow oracle extractors and keeps broad Python oracle expansion out of scope.
- Placeholder scan: The first probe-backed slice is now concrete and names the exact Go files and docs touched.
- Type consistency: `ProbeResult`, `summary_rows`, and `default_output_path` are the stable harness API used by tests and CLI.
