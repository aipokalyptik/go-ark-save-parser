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
- Property parsing for primitives, object references, byte/enum values, generic
  structs, struct arrays, simple arrays, simple maps, and simple sets.
- Read-only General, Player/Tribe local-file, Dino, Structure, Base, Stackable,
  Equipment, and save-info JSON API wrappers.
- Local cluster archive discovery and metadata loading for extensionless local
  cluster files.
- `arksave inspect`, `parse`, `players`, `tribes`, and `export-json` commands.
- Private Python oracle setup and gated private `.ark` integration test.

Still in progress:

- Full dynamic property parity for dedicated struct readers and legacy embedded
  data.
- Full domain models and APIs for dino stats/cryopods/pedigrees, full equipment
  stats, parsed player/tribe properties, local cluster item/dino payloads, full
  bases, and model-specific JSON export.
- Mutation APIs. These will remain experimental and live-server-unverified.

## Scope

Supported target inputs:

- Local `.ark` map saves.
- Local `.arkprofile` player saves.
- Local `.arktribe` tribe saves.
- Extensionless local cluster archive files where present.

Not yet supported:

- Legacy `.arktributetribe` local tribute archive parsing.

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

Export save metadata and object classes to JSON:

```sh
./bin/arksave export-json /path/to/Valguero_WP.ark /tmp/save_info.json
```

Inspect local profile and tribe archive metadata:

```sh
./bin/arksave players /path/to/76561198000000000.arkprofile
./bin/arksave tribes /path/to/123456789.arktribe
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
