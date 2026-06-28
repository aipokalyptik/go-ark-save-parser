# go-ark-save-parser

Offline Go port of
[`ark-save-parser`](https://github.com/VincentHenauGithub/ark-save-parser) for
reading Ark Survival Ascended save data from local files.

The goal is a portable, statically compiled base for command-line tools and Go
libraries that inspect `.ark`, `.arkprofile`, `.arktribe`, local cluster data,
and local tribute index files without requiring Python, virtual environments,
FTP, RCON, or live server access.

## Progress Tracking

The up-front stable task inventory is
[`docs/task-inventory.md`](docs/task-inventory.md). The detailed monitorable
task ledger is [`docs/project-task-ledger.md`](docs/project-task-ledger.md).
Together they are the source of truth for phase status, open tasks, blocked
tasks, and verification commands.

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
- Stackable APIs expose explicit typed stackable items while preserving the
  original inventory-item compatibility methods.
- Domain JSON exports include `fault_count` when fault-tolerant scans preserve
  valid rows while skipping malformed matching objects.
- Save-contained player and tribe parsing for game-table
  `PrimalPlayerDataBP`/`PrimalTribeData` objects and embedded
  `GameModeCustomBytes` player/tribe archives.
- Fault-tolerant read paths for dino, structure, equipment, stackable, and base
  scans that preserve valid parsed objects while reporting per-object parse
  failures.
- Modern cryopod dino/status payload extraction for parsed `CustomItemDatas`
  byte arrays. Empty cryopods are ignored, successfully parsed cryopodded dinos
  are returned from the Dino API keyed by cryopod item UUID, and unsupported
  embedded data can be reported through the fault-tolerant path.
- Modern cryopod saddle extraction through `DinoAPI.SaddlesFromCryopods`,
  keyed by containing cryopod item UUID when the embedded saddle has no
  independent object UUID. Equipment domain JSON includes these modern cryopod
  saddles and marks them with `in_cryopod`.
- Local cluster archive discovery plus read-only item/dino upload payload
  summaries for extensionless local cluster files. Uploaded item summaries
  include type classification, blueprint, quantity, rating, quality, and
  crafter metadata where present. Uploaded dino summaries include a
  `parse_error` field when embedded dino archive bytes cannot yet be parsed.
- Typed local cluster API helpers provide uploaded item type counts, enum-based
  item filters, dino parse-status filters, version/parse helpers, typed
  item/dino projections, embedded dino component class summaries, and summary
  metadata for library callers.
- Local `.arktributetribe` / `.arktributetribetribe` tribute index parsing
  plus JSON summaries for files and directories.
- `arksave inspect`, `parse`, `map-summary`, `object-classes`, `object-summary`,
  `property-positions`, `class-lookup`, `class-property-summary`,
  `property-filter`, `structure-health`, `structure-owner-count`,
  `structure-owners`, `structure-owner-locations`, `structure-heatmap`,
  `base-components`,
  `dinos`, `dino-wild-tamables`, `dino-babies`, `dino-best-stat`,
  `dino-most-mutated`, `dino-wild-tamed`, `equipment-summary`,
  `equipment-saddles`, `equipment-best`, `equipment-rank`,
  `equipment-owned-by`, `stackables`, `stackable-owned-by`,
  `player-inventories`, `player-roster`,
  `tribe-roster`, `player-tribe-links`, `players`, `tribes`, `cluster`,
  `cluster-summary`, `tribute`, `export-json`, `export-domain-json`,
  `export-cluster-json`, and `export-tribute-json` commands.
- Go-only provided-data E2E smoke tests for selected read-only APIs, CLI
  commands, and examples, runnable with `ARK_E2E_SAVE` or `ARK_E2E_SAVE_DIR`.
- Private Python oracle setup and gated private `.ark` integration test.
- Private Python oracle comparison for implemented offline Go examples, currently
  covering forty-six aggregate read-only and utility cases.

Still in progress:

- Full dynamic property parity for dedicated struct readers and legacy embedded
  data.
- Full domain models and APIs for legacy/modded cryopod variants, legacy/modded
  saddle payloads and cosmetics inside cryopods, full pedigree rendering/tree
  exports, richer local cluster item/dino domain models, and remaining
  model-specific JSON export edges.
- High-level semantic mutation APIs for trait/stat/growth authoring, generated
  inventory contents, and live-server-validated edits. Existing mutation
  helpers are structural copied-save workflows and remain experimental and
  live-server-unverified.

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

Run a fault-tolerant full-object parse smoke check:

```sh
./bin/arksave parse /path/to/Valguero_WP.ark
```

Summarize local map/save metadata:

```sh
./bin/arksave map-summary /path/to/Valguero_WP.ark
```

Look up save objects by class substring:

```sh
./bin/arksave object-classes /path/to/Valguero_WP.ark
./bin/arksave object-summary /path/to/Valguero_WP.ark 00112233-4455-6677-8899-aabbccddeeff
./bin/arksave property-positions /path/to/Valguero_WP.ark 00112233-4455-6677-8899-aabbccddeeff
./bin/arksave class-lookup /path/to/Valguero_WP.ark PrimalStructure
./bin/arksave class-property-summary /path/to/Valguero_WP.ark PrimalStructure
./bin/arksave property-filter /path/to/Valguero_WP.ark Health MaxHealth
```

Summarize structure health with a selected-property scan:

```sh
./bin/arksave structure-health /path/to/Valguero_WP.ark
./bin/arksave --redact structure-owner-count /path/to/Valguero_WP.ark 555
./bin/arksave structure-owners /path/to/Valguero_WP.ark
./bin/arksave --redact structure-owner-locations /path/to/Valguero_WP.ark Valguero 1
./bin/arksave structure-heatmap /path/to/Valguero_WP.ark /tmp/structure-heatmap.json 100 1
./bin/arksave base-components /path/to/Valguero_WP.ark
./bin/arksave dinos /path/to/Valguero_WP.ark
./bin/arksave dino-wild-tamables /path/to/Valguero_WP.ark
./bin/arksave dino-babies /path/to/Valguero_WP.ark
./bin/arksave dino-best-stat /path/to/Valguero_WP.ark
./bin/arksave dino-most-mutated /path/to/Valguero_WP.ark
./bin/arksave dino-wild-tamed /path/to/Valguero_WP.ark
./bin/arksave equipment-summary /path/to/Valguero_WP.ark
./bin/arksave equipment-saddles /path/to/Valguero_WP.ark
./bin/arksave equipment-best /path/to/Valguero_WP.ark
./bin/arksave equipment-rank /path/to/Valguero_WP.ark
./bin/arksave --redact equipment-owned-by /path/to/Valguero_WP.ark "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'" 555
./bin/arksave stackables /path/to/Valguero_WP.ark
./bin/arksave --redact stackable-owned-by /path/to/Valguero_WP.ark "Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'" 555
./bin/arksave player-inventories /path/to/Valguero_WP.ark
./bin/arksave player-roster /path/to/Valguero_WP.ark
./bin/arksave tribe-roster /path/to/Valguero_WP.ark
./bin/arksave player-tribe-links /path/to/save-directory
```

Export save metadata and object classes to JSON:

```sh
./bin/arksave export-json /path/to/Valguero_WP.ark /tmp/save_info.json
```

Export local cluster upload summaries to JSON from a single cluster file or a
directory of local cluster files:

```sh
./bin/arksave export-cluster-json /path/to/EOS_abc123 /tmp/cluster.json
./bin/arksave export-cluster-json /path/to/cluster-directory /tmp/clusters.json
```

Export local tribute indexes to JSON from a single tribute file or directory:

```sh
./bin/arksave export-tribute-json /path/to/abc.arktributetribe /tmp/tribute.json
./bin/arksave export-tribute-json /path/to/tribute-directory /tmp/tributes.json
```

Export implemented domain summaries to JSON:

```sh
./bin/arksave export-domain-json /path/to/Valguero_WP.ark dinos /tmp/dinos.json
./bin/arksave export-domain-json /path/to/Valguero_WP.ark structures /tmp/structures.json
./bin/arksave export-domain-json /path/to/Valguero_WP.ark equipment /tmp/equipment.json
./bin/arksave export-domain-json /path/to/Valguero_WP.ark stackables /tmp/stackables.json
./bin/arksave export-domain-json /path/to/Valguero_WP.ark players /tmp/players.json
./bin/arksave export-domain-json /path/to/Valguero_WP.ark tribes /tmp/tribes.json
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

Inspect local profile and tribe archive metadata, or point at a save directory
to aggregate all local `.arkprofile` / `.arktribe` files in it. Player directory
summaries include deaths, levels, experience, engram points, and unique unlocked
engram blueprint counts:

```sh
./bin/arksave players /path/to/76561198000000000.arkprofile
./bin/arksave players /path/to/save-directory
./bin/arksave tribes /path/to/123456789.arktribe
./bin/arksave tribes /path/to/save-directory
```

Inspect local cluster uploads:

```sh
./bin/arksave cluster /path/to/cluster-directory-or-file
./bin/arksave cluster-summary /path/to/cluster-directory-or-file
```

Inspect local tribute indexes:

```sh
./bin/arksave tribute /path/to/tribute-directory-or-file
```

Create an experimental mutation copy, remove objects from a copied save, upsert
object bytes, or upsert a custom-table value from hex bytes:

```sh
./bin/arksave mutate copy /path/to/Valguero_WP.ark /tmp/Valguero_copy.ark
./bin/arksave mutate remove-object /path/to/Valguero_WP.ark /tmp/Valguero_removed.ark 00112233-4455-6677-8899-aabbccddeeff
./bin/arksave mutate remove-class-contains /path/to/Valguero_WP.ark /tmp/Valguero_no_spyglass.ark SuperSpyglass
./bin/arksave mutate import-base-binary /path/to/Valguero_WP.ark /tmp/Valguero_base_import.ark /tmp/base-export
./bin/arksave mutate import-structure-binary /path/to/Valguero_WP.ark /tmp/Valguero_structure_import.ark /tmp/structure-export
./bin/arksave mutate import-dino-binary /path/to/Valguero_WP.ark /tmp/Valguero_dino_import.ark /tmp/dino-export
./bin/arksave mutate import-equipment-binary /path/to/Valguero_WP.ark /tmp/Valguero_equipment_import.ark /tmp/equipment-export
./bin/arksave mutate put-object-hex /path/to/Valguero_WP.ark /tmp/Valguero_object.ark 00112233-4455-6677-8899-aabbccddeeff 090807
./bin/arksave mutate replace-object-property-hex /path/to/Valguero_WP.ark /tmp/Valguero_property.ark 00112233-4455-6677-8899-aabbccddeeff DinoID1 0 010203
./bin/arksave mutate put-custom /path/to/Valguero_WP.ark /tmp/Valguero_custom.ark Extra 090807
```

Run tests:

```sh
make test
```

Run Go-only read-only E2E smoke tests against private/provided save data:

```sh
ARK_E2E_SAVE=/absolute/path/to/Valguero_WP.ark make e2e-test
ARK_E2E_SAVE_DIR=/absolute/path/to/SavedArks make e2e-test
```

The E2E path opens local `.ark` data, exercises selected read-only APIs, CLI
commands, local tribute handling, and aggregate-output examples, writes CLI JSON
only to temporary test directories, and does not write private output. Python
oracle expansion is not required for this validation path.

Run the private oracle integration test:

```sh
ARK_ORACLE_SAVE=/absolute/path/to/save.ark make oracle-test
ARK_ORACLE_SAVE=/absolute/path/to/save.ark ARK_ORACLE_TRIBUTE=/absolute/path/to/file.arktributetribe make oracle-test
```

Run the private oracle comparison suite for implemented Go examples:

```sh
ARK_ORACLE_SAVE=/absolute/path/to/save.ark make oracle-compare
```

Run a focused private oracle comparison without overwriting the committed
aggregate summary:

```sh
ARK_ORACLE_SAVE=/absolute/path/to/save.ark .oracle/venv/bin/python scripts/oracle_compare.py --case dino_heatmap
```

Run standalone Go examples:

```sh
go run ./examples/map_summary /path/to/Valguero_WP.ark
go run ./examples/object_classes /path/to/Valguero_WP.ark
go run ./examples/property_filter /path/to/Valguero_WP.ark TamerString Health
go run ./examples/dino_best_stat --no-cryos /path/to/Valguero_WP.ark
go run ./examples/stackable_count /path/to/Valguero_WP.ark /Game/Path/PrimalItemResource_Example.PrimalItemResource_Example_C
go run ./examples/equipment_best /path/to/Valguero_WP.ark
go run ./examples/player_list /path/to/save-directory
go run ./examples/player_and_tribe_data /path/to/save-directory
go run ./examples/player_inventories /path/to/Valguero_WP.ark
go run ./examples/player_unlocked_engrams /path/to/save-directory
go run ./examples/equipment_rank /path/to/Valguero_WP.ark
go run ./examples/base_export_from_save /path/to/Valguero_WP.ark /tmp/base-export
go run ./examples/structure_export_from_save /path/to/Valguero_WP.ark /tmp/structure-export
go run ./examples/dino_export_from_save /path/to/Valguero_WP.ark /tmp/dino-export
go run ./examples/equipment_export_from_save /path/to/Valguero_WP.ark /tmp/equipment-export
go run ./examples/local_profiles /path/to/save-directory
go run ./examples/cluster_json /path/to/EOS_abc123
go run ./examples/cluster_typed /path/to/EOS_abc123
go run ./examples/local_tribute /path/to/tribute-directory-or-file
go run ./examples/tribute_json /path/to/tribute-directory-or-file
go run ./examples/mutation_copy /path/to/Valguero_WP.ark /tmp/Valguero_copy.ark
go run ./examples/mutation_copy remove-class-contains /path/to/Valguero_WP.ark /tmp/Valguero_no_spyglass.ark SuperSpyglass
go run ./examples/mutation_copy import-base-binary /path/to/Valguero_WP.ark /tmp/Valguero_base_import.ark /tmp/base-export
go run ./examples/mutation_copy import-structure-binary /path/to/Valguero_WP.ark /tmp/Valguero_structure_import.ark /tmp/structure-export
go run ./examples/mutation_copy import-dino-binary /path/to/Valguero_WP.ark /tmp/Valguero_dino_import.ark /tmp/dino-export
go run ./examples/mutation_copy import-equipment-binary /path/to/Valguero_WP.ark /tmp/Valguero_equipment_import.ark /tmp/equipment-export
go run ./examples/mutation_copy put-object-hex /path/to/Valguero_WP.ark /tmp/Valguero_object.ark 00112233-4455-6677-8899-aabbccddeeff 090807
go run ./examples/mutation_copy replace-object-property-hex /path/to/Valguero_WP.ark /tmp/Valguero_property.ark 00112233-4455-6677-8899-aabbccddeeff DinoID1 0 010203
go run ./examples/mutation_copy put-custom /path/to/Valguero_WP.ark /tmp/Valguero_custom.ark Extra 090807
```

## Mutation Safety

Mutation helpers never modify the input file in place and always require a new
output path. Generated mutation copies are written with `0600` permissions.
They are structurally tested by reopening and reparsing the copied SQLite save,
but live Ark server behavior is unverified; command output labels these files as
experimental and live-server-unverified. Raw property replacement expects the
full encoded property record for the target save's name table; it is a
structural escape hatch for copied-save workflows, not a high-level semantic
trait/stat/growth authoring API.

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
