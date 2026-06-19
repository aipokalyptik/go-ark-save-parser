# go-ark-save-parser

Offline Go port of
[`ark-save-parser`](https://github.com/VincentHenauGithub/ark-save-parser) for
reading Ark Survival Ascended save data from local files.

The goal is a portable, statically compiled base for command-line tools and Go
libraries that inspect `.ark`, `.arkprofile`, `.arktribe`, local cluster data,
and local tribute index files without requiring Python, virtual environments,
FTP, RCON, or live server access.

## Current Status

Implemented:

- Primitive ARK binary reader for little-endian values, ARK strings, UUIDs,
  name-table lookups, zlib inflation, and wildcard decompression.
- Local SQLite-backed `.ark` opening through pure-Go SQLite.
- Save header parsing, name table loading, custom value reads, object ID
  enumeration, raw object reads, class-name lookup, and generic object parsing.
- Property parsing for primitives, object references, soft object/name values,
  byte/enum values, generic structs, struct arrays, simple arrays, simple maps,
  and simple sets.
- Read-only General, Player/Tribe local-file, local tribute, Dino, Structure,
  Base, Stackable, Equipment, save-info JSON, and domain JSON API wrappers.
- Local cluster archive discovery plus read-only item/dino upload payload
  summaries for extensionless local cluster files.
- Local `.arktributetribe` / `.arktributetribetribe` tribute index parsing.
- `arksave inspect`, `parse`, `players`, `tribes`, `cluster`, `tribute`,
  `export-json`, `export-domain-json`, and `export-cluster-json` commands.
- Private Python oracle setup and gated private `.ark` integration test.

Still in progress:

- Full dynamic property parity for dedicated struct readers and legacy embedded
  data.
- Full domain models and APIs for dino stats/cryopods/pedigrees, full equipment
  stats, parsed player/tribe properties, richer local cluster item/dino domain
  models, full bases, and model-specific JSON export.
- Mutation APIs beyond copy/remove/upsert structural helpers. All mutation
  helpers remain experimental and live-server-unverified.

## Scope

Supported target inputs:

- Local `.ark` map saves.
- Local `.arkprofile` player saves.
- Local `.arktribe` tribe saves.
- Extensionless local cluster archive files where present.
- Local `.arktributetribe` / `.arktributetribetribe` tribute index files.

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

Export local cluster upload summaries to JSON:

```sh
./bin/arksave export-cluster-json /path/to/EOS_abc123 /tmp/cluster.json
```

Export implemented domain summaries to JSON:

```sh
./bin/arksave export-domain-json /path/to/Valguero_WP.ark dinos /tmp/dinos.json
./bin/arksave export-domain-json /path/to/Valguero_WP.ark structures /tmp/structures.json
./bin/arksave export-domain-json /path/to/Valguero_WP.ark equipment /tmp/equipment.json
./bin/arksave export-domain-json /path/to/Valguero_WP.ark stackables /tmp/stackables.json
./bin/arksave export-domain-json /path/to/Valguero_WP.ark bases /tmp/bases.json
```

Use `--redact` when command output or export files need to be safe for logs,
issue comments, or aggregate reporting:

```sh
./bin/arksave --redact players /path/to/76561198000000000.arkprofile
./bin/arksave --redact export-json /path/to/Valguero_WP.ark /tmp/save_info.redacted.json
./bin/arksave --redact export-domain-json /path/to/Valguero_WP.ark dinos /tmp/dinos.redacted.json
./bin/arksave --redact export-cluster-json /path/to/EOS_abc123 /tmp/cluster.redacted.json
```

Redacted summaries keep counts and versions but hide local paths, profile/tribe
names, IDs, archive class lists, cluster upload details, and object/domain item
details.

Profile, tribe, and local cluster archive readers apply a default 512 MiB
stat-before-read limit. Library callers that intentionally handle larger local
archives can use the `Open...WithOptions` APIs in `arkprofile` and `arkcluster`
to set a different limit. Zlib inflation also has an explicit bounded helper
for callers that need to guard compressed payload expansion.

CLI summaries and JSON exports are save-derived operational data. They can
include local paths, class names, object IDs, player or tribe identifiers,
locations, crafter names, and cluster upload identifiers depending on the
command. JSON export files are written with `0600` permissions by default, but
you should still treat them as private and avoid committing or sharing them
unless they have been reviewed and sanitized.

Inspect local profile and tribe archive metadata:

```sh
./bin/arksave players /path/to/76561198000000000.arkprofile
./bin/arksave tribes /path/to/123456789.arktribe
```

Inspect local cluster uploads:

```sh
./bin/arksave cluster /path/to/cluster-directory-or-file
```

Inspect local tribute indexes:

```sh
./bin/arksave tribute /path/to/tribute-directory-or-file
```

Create an experimental mutation copy or remove an object from a copied save:

```sh
./bin/arksave mutate copy /path/to/Valguero_WP.ark /tmp/Valguero_copy.ark
./bin/arksave mutate remove-object /path/to/Valguero_WP.ark /tmp/Valguero_removed.ark 00112233-4455-6677-8899-aabbccddeeff
```

Run tests:

```sh
make test
```

Run the private oracle integration test:

```sh
ARK_ORACLE_SAVE=/absolute/path/to/save.ark make oracle-test
ARK_ORACLE_SAVE=/absolute/path/to/save.ark ARK_ORACLE_TRIBUTE=/absolute/path/to/file.arktributetribe make oracle-test
```

Run standalone Go examples:

```sh
go run ./examples/map_summary /path/to/Valguero_WP.ark
go run ./examples/object_classes /path/to/Valguero_WP.ark
go run ./examples/local_profiles /path/to/save-directory
go run ./examples/cluster_json /path/to/EOS_abc123
go run ./examples/local_tribute /path/to/tribute-directory-or-file
go run ./examples/mutation_copy /path/to/Valguero_WP.ark /tmp/Valguero_copy.ark
```

## Mutation Safety

Mutation helpers never modify the input file in place and always require a new
output path. Generated mutation copies are written with `0600` permissions.
They are structurally tested by reopening and reparsing the copied SQLite save,
but live Ark server behavior is unverified; command output labels these files as
experimental and live-server-unverified.

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
`docs/development.md`, `docs/phase-1-report.md`, `docs/oracle-summary.md`, and
`docs/oracle-comparison-summary.md` for regeneration instructions,
commit-safe details, and implemented-example comparison status.
