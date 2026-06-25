# Phase 3 Idiomatic Go Refactor

Phase 3 is complete when the translated implementation is organized into stable
Go packages without losing oracle parity.

## Current Package Shape

- `arkbinary`: primitive binary reader/decompression/name-table handling.
- `arksave`: local SQLite-backed `.ark` access and save metadata.
- `arkproperty`: dynamic property parsing.
- `arkobject`: generic game object parsing.
- `arkprofile`: local `.arkprofile` and `.arktribe` archive readers.
- `arkcluster`: local extensionless cluster file discovery and payload parsing.
- `arkapi`: high-level offline query APIs.
- `arkmutation`: experimental explicit-output copy/mutation helpers.
- `cmd/arksave`: offline CLI.

## Remaining Refactor Tasks

- [x] Expand `arkproperty` beyond primitives: structs, arrays, maps, sets,
      object references, unknown struct fallback, and legacy parser isolation.
- [ ] Split domain models under `arkobject` or subpackages for dino, structure,
      equipment, stackable, player, tribe, inventory, and local cluster data.
- [x] Add first typed local player/tribe lookup layer for player data IDs, tribe
      IDs, parsed player summaries, and parsed tribe detail objects.
- [x] Add typed dino stat and mutation selection helpers for best-stat and
      most-mutated-tamed read-only workflows.
- [x] Expand typed dino best-stat selection with class, tame-state, stat-list,
      scope, and level-cap options.
- [x] Add typed structure inventory-container lookup and stackable owner filter
      helpers for upstream owner-count example flows.
- [x] Add typed equipment owner filter helpers through structure inventory
      containers for upstream owner-of-items workflows.
- [x] Add typed equipment crafted and average-stat ranking helpers for upstream
      equipment ranking workflows.
- [x] Surface equipment crafted and average-stat ranking helpers through domain
      JSON export for CLI/library consumers.
- [ ] Add typed API layers for full dino, full structure, equipment, full
      stackable, base, additional model-specific JSON export, local cluster
      domain modeling, and remaining player/tribe upstream parity.
- [ ] Replace duplicated synthetic fixture builders in tests with internal test
      helpers. `internal/testfixtures` now centralizes public synthetic SQLite
      saves, generic object payloads, local profile/tribe/cluster archive
      payloads, compact local tribute indexes, shared name-table-ID property
      writers, shared object header/terminator wrapping, and shared
      sparse-file/max-size fixtures plus header/string/property encoding for
      examples, CLI tests, `arkprofile`, `arkapi`, `arkarchive`, `arkcluster`,
      `arktribute`, and benchmarks. Structure, base, stackable, and equipment
      synthetic object builders now use the shared object wrapper; remaining
      lower-level domain-specific parser fixtures still need incremental
      migration.
- [x] Route `arkapi` synthetic save fixtures through `internal/testfixtures`
      instead of repeated direct SQLite table creation in each domain test.
- [x] Add benchmarks for full save open/object enumeration, object parse, query
      filters, and JSON export.
- [x] Keep `inspect` as metadata-only and make `parse` perform a
      fault-tolerant full-object parse summary, then expand `cmd/arksave`
      with local profile/tribe file and directory summaries plus
      save-info/domain JSON exports.
- [x] Keep mutation helpers in an explicit experimental surface that requires an
      output path, including copy, object removal, object-byte upsert, and
      custom-table upsert CLI commands.
- [x] Re-run oracle integration after major parser/API expansion.
