# Phase 3 Idiomatic Go Refactor

Phase 3 is complete when the translated implementation is organized into stable
Go packages without losing oracle parity.

## Current Package Shape

- `arkbinary`: primitive binary reader/decompression/name-table handling.
- `arksave`: local SQLite-backed `.ark` access and save metadata.
- `arkproperty`: dynamic property parsing.
- `arkobject`: generic game object parsing.
- `arkapi`: high-level offline query APIs.
- `cmd/arksave`: offline CLI.

## Remaining Refactor Tasks

- [ ] Expand `arkproperty` beyond primitives: structs, arrays, maps, sets,
      object references, unknown struct fallback, and legacy parser isolation.
- [ ] Split domain models under `arkobject` or subpackages for dino, structure,
      equipment, stackable, player, tribe, inventory, and local cluster data.
- [ ] Add typed API layers for player/tribe, dino, full structure, equipment,
      full stackable, base, model-specific JSON export, and local cluster parsing.
- [ ] Replace duplicated synthetic fixture builders in tests with internal test
      helpers.
- [ ] Add benchmarks for full save open/object enumeration, object parse, query
      filters, and JSON export.
- [x] Expand `cmd/arksave` commands beyond `inspect`/`parse` with local
      profile/tribe metadata and save-info JSON export.
- [ ] Keep mutation helpers in an explicit experimental surface that requires an
      output path.
- [ ] Re-run oracle integration after each major parser expansion.
