# Phase 1 Oracle Setup

Phase 1 is complete when Python oracle behavior is reproducible from private
backup data for the selected offline parity features, and the reference oracle
work that was already run has a recorded, sanitized status. Expanding the
Python codebase or oracle suite for every feasible upstream example is no
longer a project goal; new coverage should be added as Go tests, Go examples,
and provided-data E2E paths.

For the cross-phase monitorable checklist, see
[`docs/project-task-ledger.md`](project-task-ledger.md).

## Current Status

Phase 1 is closed for the selected offline parity scope. The committed
commit-safe evidence is the aggregate backup summary, upstream test/testbench
status, example classification, oracle comparison summary, and Phase 1 report.
Private raw logs, manifests, paths, and outputs remain under `.oracle/`.

Do not add new Python oracle examples or improve upstream Python code as normal
project work. Future oracle interaction should be limited to narrow checks that
directly unblock Phase 2 Go parity.

## Requirements

- Keep all private save files, manifests, raw output, snapshots, and debug dumps
  under `.oracle/`.
- Do not commit raw paths, player names, tribe names, object identifiers, or
  other private save contents.
- Use upstream `ark-save-parser` commit
  `4f7cc91fb96a080321bfbc884ba81bd897f72c49` as the oracle source.
- Use Python 3.13+ for oracle setup because the local system Python 3.9.6 cannot
  satisfy current upstream dependency versions.
- Treat upstream packaged tests as blocked if their private `tests/test_data`
  corpus is absent.
- Use upstream `testbench/` for arbitrary local `.ark` saves.
- Rewrite example inputs to local files only when using existing oracle
  references for selected-feature parity.
- Skip FTP and RCON examples as intentionally out of scope.
- Run mutation examples only against throwaway copies and classify them as
  structurally tested, not live-server-verified.

## Task Ledger

- [x] Create repository hygiene and public GitHub remote.
- [x] Verify and extract `~/Downloads/SavedArks.tar.bz2` into `.oracle/data`.
- [x] Clone upstream into `.oracle/upstream` at the pinned commit.
- [x] Build `.oracle/venv` with a suitable Python runtime.
- [x] Generate `.oracle/manifest.json` with private file details.
- [x] Commit a sanitized backup summary with counts by file type only.
- [x] Run upstream packaged tests and record sanitized status.
- [x] Run upstream `testbench/` for each usable `.ark` save and record sanitized
      status.
- [x] Classify all offline-compatible examples.
- [x] Add and run a privacy-safe comparison harness for implemented Go example
      counterparts; current committed aggregate status is in
      `docs/oracle-comparison-summary.md`, with detailed values kept under
      `.oracle/output/oracle-comparison.json`.
- [x] Cover selected read-only examples with Go counterparts and privacy-safe
      aggregate comparisons where existing oracle evidence is useful.
- [x] Represent selected mutation examples with copy/write/reopen structural
      Go tests where feasible; live-server behavior remains unverified by
      design.
- [x] Review privacy boundaries and oracle completeness.
