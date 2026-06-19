# Development Guide

This project is an offline-only Go port of upstream `ark-save-parser` at commit
`4f7cc91fb96a080321bfbc884ba81bd897f72c49`.

## Local Verification

Use these commands before committing parser, API, CLI, or docs changes:

```sh
go test ./...
make build
```

Use benchmarks when changing parser or query paths:

```sh
make bench
```

The build target uses `CGO_ENABLED=0` so the CLI remains portable and does not
depend on a system SQLite library.

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

Intentionally unsupported:

- FTP
- RCON
- live server control or validation

Legacy `.arktributetribe` parsing remains unsupported until a local-file oracle
path proves enough format behavior to port safely.
