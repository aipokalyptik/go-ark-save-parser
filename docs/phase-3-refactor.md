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
- [ ] Add typed API layers for player/tribe, dino, full structure, equipment,
      full stackable, base, additional model-specific JSON export, and local
      cluster domain modeling.
- [ ] Replace duplicated synthetic fixture builders in tests with internal test
      helpers. `internal/testfixtures` now centralizes public synthetic SQLite
      saves, generic object payloads, local profile/tribe/cluster archive
      payloads, and shared header/string/property encoding for examples, CLI
      tests, `arkprofile`, `arkapi`, and benchmarks; lower-level
      domain-specific parser fixtures still need incremental migration.
- [x] Add benchmarks for full save open/object enumeration, object parse, query
      filters, and JSON export.
- [x] Expand `cmd/arksave` commands beyond `inspect`/`parse` with local
      profile/tribe metadata and save-info JSON export.
- [x] Keep mutation helpers in an explicit experimental surface that requires an
      output path.
- [x] Re-run oracle integration after major parser/API expansion.
