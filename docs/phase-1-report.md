# Phase 1 Report

Phase 1 establishes the Python oracle and private-data handling for the offline
Go port.

## Completed Setup

- Public GitHub repository created:
  `https://github.com/aipokalyptik/go-ark-save-parser`
- Main branch tracks `origin/main`.
- Go module initialized as `github.com/aipokalyptik/go-ark-save-parser`.
- Private oracle data is ignored under `.oracle/`.
- Upstream source is cloned privately under `.oracle/upstream` and pinned to
  commit `4f7cc91fb96a080321bfbc884ba81bd897f72c49`.
- Python 3.13.14 was installed through Homebrew for oracle execution.
- Upstream package and pytest are installed in `.oracle/venv`.
- Private backup tarball `~/Downloads/SavedArks.tar.bz2` was extracted under
  `.oracle/data`.

## Private Backup Coverage

See [oracle-summary.md](oracle-summary.md) for the commit-safe aggregate
inventory.

The backup contains:

- One map save.
- Hundreds of player profiles.
- Hundreds of tribe saves.
- Local tribute/cluster-like files.

## Upstream Fixed Tests

Upstream packaged tests are blocked in the public checkout because
`tests/test_data` is absent. Running the suite confirms it errors on missing
fixtures such as fixed Ragnarok map saves before parser behavior can be judged.

Detailed raw output is private in `.oracle/output/upstream-pytest.log`.

## Arbitrary-Save Testbench

The upstream `testbench/` suite is the primary Phase 1 oracle because it accepts
an arbitrary `.ark` save and records snapshot metrics for that save.

Current status: partially successful against the supplied private map save.

- Runtime: 422.40 seconds.
- Result: 5 passed, 1 failed, stopped after first failure.
- Core save parsing succeeded.
- Parsed object count: 2,164,851.
- Faulty object count: 0.
- The failure occurred in `tests/test_02_dinos.py::test_get_all` while parsing
  legacy embedded cryopod dino data from modded content.
- The failing upstream path expected a 16-byte zero block in a legacy
  `CustomDinoData` struct and found non-zero bytes instead.

This means the live backup is a useful oracle for core `.ark` loading, name
tables, object enumeration, game time, and general object access. Full Dino API
parity must treat modded legacy cryopod parsing as a known upstream failure
case unless the Go port intentionally improves on upstream behavior.

Raw output is private in `.oracle/output/testbench-pytest.log`.

## Example Classification

See [upstream-oracle-classification.md](upstream-oracle-classification.md).

The important execution rules are:

- Read-only examples can be patched to local paths.
- Export and heatmap examples write output only under `.oracle/output`.
- Mutation examples run only on throwaway copies.
- FTP and RCON examples are intentionally skipped.

## Privacy Review

Committed files include only aggregate counts and classification notes. Private
paths, hashes, snapshots, raw outputs, debug dumps, object bytes, player names,
tribe names, and object identifiers remain under `.oracle/` and are ignored by
git.

## Go Smoke Verification

The Go CLI built with `CGO_ENABLED=0` successfully inspected the private map save
through `bin/arksave inspect`.

Commit-safe observed values:

- Save version: 14.
- Object count: 2,164,851.
- The object count matches the upstream Python testbench core parse result.
