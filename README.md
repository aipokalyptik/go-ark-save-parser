# go-ark-save-parser

Offline Go port of
[`ark-save-parser`](https://github.com/VincentHenauGithub/ark-save-parser) for
reading Ark Survival Ascended save data from local files.

The goal is a portable, statically compiled base for command-line tools and Go
libraries that inspect `.ark`, `.arkprofile`, `.arktribe`, and local cluster data
without requiring Python, virtual environments, FTP, RCON, or live server access.

## Current Status

Implemented:

- Primitive ARK binary reader for little-endian values, ARK strings, UUIDs,
  name-table lookups, zlib inflation, and wildcard decompression.
- Local SQLite-backed `.ark` opening through pure-Go SQLite.
- Save header parsing, name table loading, custom value reads, object ID
  enumeration, raw object reads, class-name lookup, and generic object parsing.
- Primitive property parsing for `BoolProperty`, `IntProperty`, `StrProperty`,
  and `None` termination.
- Read-only General API wrapper.
- Initial `arksave inspect` / `arksave parse` CLI summary command.
- Private Python oracle setup and gated private `.ark` integration test.

Still in progress:

- Full dynamic property parity for structs, arrays, maps, sets, object
  references, and legacy embedded data.
- Domain models and APIs for dinos, structures, equipment, stackables, players,
  tribes, local cluster files, bases, and JSON export.
- Mutation APIs. These will remain experimental and live-server-unverified.

## Scope

Supported target inputs:

- Local `.ark` map saves.
- Local `.arkprofile` player saves.
- Local `.arktribe` tribe saves.
- Local cluster files where present.

Out of scope:

- FTP.
- RCON.
- Live server integration.
- Automated proof that modified saves load correctly in a running Ark server.

## Usage

Build the CLI:

```sh
make build
```

Inspect a local save:

```sh
./bin/arksave inspect /path/to/Valguero_WP.ark
```

Run tests:

```sh
make test
```

Run the private oracle integration test:

```sh
ARK_ORACLE_SAVE=/absolute/path/to/save.ark make oracle-test
```

## Library Sketch

```go
save, err := arksave.Open("/path/to/Valguero_WP.ark")
if err != nil {
    return err
}
defer save.Close()

ids, err := save.ObjectIDs()
if err != nil {
    return err
}

obj, err := save.Object(ids[0])
```

## Private Oracle

Private save data and oracle output live under `.oracle/` and are ignored by git.
The current oracle was generated from `~/Downloads/SavedArks.tar.bz2`; see
`docs/phase-1-report.md` and `docs/oracle-summary.md` for commit-safe details.
