# Development Guide

This project is an offline-only Go port of upstream `ark-save-parser` at commit
`4f7cc91fb96a080321bfbc884ba81bd897f72c49`.

## Progress Tracking

Use [`docs/task-inventory.md`](task-inventory.md) for stable task IDs and the
up-front goal checklist. Update [`docs/project-task-ledger.md`](project-task-ledger.md)
in the same commit whenever a phase task, blocker, or completion status
changes. Together those files are the repository source of truth for
monitorable project progress.

## Local Verification

Use these commands before committing parser, API, CLI, or docs changes:

```sh
make verify
```

`make verify` runs the full Go test suite with repo-local build cache settings,
compiles the Python oracle helper scripts, and builds the static CLI binary.

Use benchmarks when changing parser or query paths:

```sh
make bench
```

For repeated object lookup/parse changes, compare the raw object row cache
against the uncached path:

```sh
GOCACHE=$PWD/.cache/go-build go test ./arkapi -bench 'BenchmarkObjectParse' -benchtime=1x
```

The cache is opt-in on `arksave.Save` via `SetObjectCacheEnabled(true)`.
Disable it or call `ClearObjectCache()` when reusing a save handle after
out-of-band copied-save mutations. Object row cache access is guarded for
concurrent `ObjectBinary` reads, but higher-level API concurrency is not a
compatibility guarantee unless a specific path has tests.

The build target uses `CGO_ENABLED=0` so the CLI remains portable and does not
depend on a system SQLite library.

## Provided-Data Go E2E

Prefer expanding Go tests, examples, and E2E coverage for chosen offline
features. Do not add Python oracle cases first unless an existing oracle check
is the narrowest way to answer a specific parity question.

Run the Go-only provided-data smoke test with either a direct `.ark` path or a
directory containing provided save files:

```sh
ARK_E2E_SAVE=/absolute/path/to/Valguero_WP.ark make e2e-test
ARK_E2E_SAVE_DIR=/absolute/path/to/SavedArks make e2e-test
```

`make e2e-test` is read-only. It opens the map save, checks save metadata,
exercises object enumeration and selected-property scans, runs read-only CLI
commands, and, when a directory is supplied, also exercises local player and
tribe discovery. CLI JSON output is written only to temporary test directories.
It skips cleanly when no provided data environment variables are set.

## Oracle Regeneration

Private oracle files must stay under `.oracle/`, which is ignored by git.

Expected local source:

```sh
~/Downloads/SavedArks.tar.bz2
```

Regeneration outline:

```sh
mkdir -p .oracle/data .oracle/output
tar -xjf ~/Downloads/SavedArks.tar.bz2 -C .oracle/data
git clone https://github.com/VincentHenauGithub/ark-save-parser .oracle/upstream
git -C .oracle/upstream checkout 4f7cc91fb96a080321bfbc884ba81bd897f72c49
python3.13 -m venv .oracle/venv
.oracle/venv/bin/pip install -e .oracle/upstream pytest
```

Run the committed Go oracle gate against a selected private map save:

```sh
ARK_ORACLE_SAVE=/absolute/path/to/private/save.ark make oracle-test
```

Add `ARK_ORACLE_TRIBUTE` to the same command when checking a private local
tribute index fixture:

```sh
ARK_ORACLE_SAVE=/absolute/path/to/private/save.ark ARK_ORACLE_TRIBUTE=/absolute/path/to/private/file.arktributetribe make oracle-test
```

Run the implemented example comparison harness against the Python oracle:

```sh
ARK_ORACLE_SAVE=/absolute/path/to/private/save.ark make oracle-compare
```

`oracle-compare` writes detailed private values and stdout/stderr to
`.oracle/output/oracle-comparison.json` and updates the commit-safe aggregate
status in `docs/oracle-comparison-summary.md`.

Focused oracle comparisons are useful when validating a newly added case without
rerunning every upstream example comparison. Focused output defaults to ignored
files under `.oracle/output`:

```sh
ARK_ORACLE_SAVE=/absolute/path/to/private/save.ark .oracle/venv/bin/python scripts/oracle_compare.py --case dino_heatmap
```

Upstream fixed tests require non-public `tests/test_data` and are recorded as
blocked in `docs/upstream-oracle-classification.md`. The upstream `testbench/`
suite is the useful arbitrary-save oracle path for private `.ark` files.

## Privacy Rules

Never commit:

- `.oracle/`
- raw save files or extracted save directories
- private manifests with paths, hashes, object identifiers, player names, or
  tribe names
- raw oracle output, snapshots, debug dumps, or generated private JSON/text logs
- mutation outputs created from private saves

Committed oracle documentation must stay aggregate or sanitized. The safe
inventory summary is `docs/oracle-summary.md`.

CLI and JSON output is also sensitive at runtime. Commands can print or export
local paths, object IDs, class names, player/tribe identifiers, locations,
crafter names, and cluster upload identifiers. Export files are created with
`0600` permissions, but generated outputs from private saves still belong under
`.oracle/output` or another ignored private directory unless they have been
explicitly sanitized.

Use the CLI `--redact` option for aggregate-only output intended for logs,
issues, or committed notes. Redacted mode hides stdout paths and identifiers,
omits archive class details in profile/tribe summaries, omits cluster upload
details, and writes aggregate JSON exports without object, domain item, or
cluster item/dino detail records.

Profile, tribe, and local cluster archive reads are guarded by a default 512 MiB
stat-before-read limit so accidental giant files fail before whole-file memory
allocation. Use `arkprofile.OpenPlayerProfileWithOptions`,
`arkprofile.OpenTribeSaveWithOptions`, `arkcluster.OpenWithOptions`, or
`arkcluster.OpenDirectoryWithOptions` when a tool needs a tighter or looser
local archive limit. Map `.ark` saves are SQLite-backed and are not read as a
single byte slice by the public open path.

## Adding Fixtures

Prefer synthetic fixtures in tests. If a real save is required:

1. Put it under `.oracle/`.
2. Gate the test behind an environment variable.
3. Assert sanitized counts or structural properties only.
4. Keep output under `.oracle/output`.
5. Do not paste paths, names, UUIDs, or raw object bytes into committed files.

## Mutation Work

Mutation helpers live in `arkmutation` and must:

- require an explicit output path
- never modify the input save in place
- write copied outputs with private file permissions
- structurally verify copied outputs by reopening and reparsing them
- remain documented as live-server-unverified

Automated tests can prove that a copied SQLite save reopens and parses. They do
not prove that Ark Survival Ascended will accept the modified save on a live
server.

## Offline Scope

Supported local inputs:

- `.ark`
- `.arkprofile`
- `.arktribe`
- extensionless local cluster files when present
- `.arktributetribe` and `.arktributetribetribe` local tribute index files

Intentionally unsupported:

- FTP
- RCON
- live server control or validation

Legacy archive object parsing remains unsupported where an input is not one of
the modern local archive formats or the compact local tribute index format.
