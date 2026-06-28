# Phase 3 Idiomatic Go Refactor

Phase 3 is complete when the translated implementation is organized into stable
Go packages without losing oracle parity.

For the cross-phase monitorable checklist, see
[`docs/project-task-ledger.md`](project-task-ledger.md).

## Current Status

Phase 3 is the active execution phase. Refactor slices should preserve the
closed Phase 2 behavior, keep tests passing, and avoid reopening Python oracle
expansion unless a new Go failure exposes a concrete offline parity defect.

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
- [x] Add explicit typed stackable item model and compatibility API wrappers
      over the existing inventory-item stackable behavior.
- [x] Add typed equipment owner filter helpers through structure inventory
      containers for upstream owner-of-items workflows.
- [x] Add typed equipment crafted and average-stat ranking helpers for upstream
      equipment ranking workflows.
- [x] Surface equipment crafted and average-stat ranking helpers through domain
      JSON export for CLI/library consumers.
- [x] Add first typed local cluster API wrapper for uploaded item type counts,
      dino parse-status filtering, and summary metadata.
- [x] Add typed local cluster domain projections for uploaded items and dinos
      while preserving raw `arkcluster` accessors.
- [x] Add typed local cluster uploaded-item type constants and helper methods
      while preserving string-based filter and JSON compatibility.
- [x] Add typed local cluster enum filters, upload-version helpers, parse-status
      helpers, and embedded dino component class summaries for Go callers.
- [x] Add typed local cluster uploaded-item aggregate summaries for item type,
      version support, crafted item, quantity, rating, and quality counts.
- [x] Add typed local cluster uploaded-dino aggregate summaries for parsed
      dinos, parse errors, version support, component presence, and embedded
      object counts.
- [x] Add typed local cluster uploaded item short names and uploaded dino
      primary/short class names for JSON/library/CLI consumers.
- [x] Surface typed player and tribe exports through domain JSON for CLI and
      library consumers.
- [x] Add `arkapi.NewPlayerFromPath` so examples and library consumers can
      open either save files or local save directories with explicit player or
      tribe fallback behavior.
- [x] Add `arkapi.PlayerInventorySummaryForPlayers` and
      `arkapi.PlayerInventorySummaryFromPath` so player inventory examples use
      reusable typed aggregation instead of local counting logic.
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
      `arktribute`, `arksave`, and benchmarks. Structure, base, stackable,
      equipment, and core save synthetic object builders now use the shared
      object wrapper; `arkapi` general/core synthetic save helpers now delegate
      header and object wrapping to shared fixtures; `arkprofile` malformed
      archive tests reuse shared archive framing/string/property writers;
      dino/equipment string property payload writers and equipment positioned
      UInt16 property payload writers are shared; dino scalar, object-reference,
      positioned stat/color, and name-array property payload writers use shared
      helpers; dino custom item data struct/byte-array payload writers use
      shared helpers; low-level `arkproperty` and `arkobject` name-table-ID,
      integer, float, string, and compound custom-data test writers use shared
      helpers; API malformed class-only object row fixtures use a shared helper;
      API/benchmark actor-transform custom table fixtures use a shared helper;
      save-contained player/tribe game object fixture payloads share archive
      property builders while keeping game-object framing explicit;
      embedded `GameModeCustomBytes` player/tribe fixture assembly lives in
      shared testfixtures; minimal embedded cryopod archive test payloads use a
      shared helper;
      and modern cryopod embedded dino/saddle payload builders are shared by API
      and object-model tests; save-layer malformed full-object truncation
      fixtures use a shared helper; parsed `CustomItemDatas`
      cryopod/custom-data fixtures live in `internal/propertyfixtures`, and
      binary `CustomItemDatas` writers live in shared testfixtures for API and
      object-model tests; ID-table Vector struct property writers are shared by
      example/player-location fixtures; base linked-structure
      object-reference array fixtures use shared ID-table array writers.
      Remaining lower-level domain-specific parser fixtures and non-save
      malformed object-shape fixtures still need incremental migration.
- [x] Route `arkapi` synthetic save fixtures through `internal/testfixtures`
      instead of repeated direct SQLite table creation in each domain test.
- [x] Add benchmarks for full save open/object enumeration, object parse, query
      filters, and JSON export.
- [x] Add opt-in `arksave.Save` object row cache controls for repeated object
      lookup/parse workflows, plus cached object-parse benchmark coverage.
- [x] Guard opt-in object row cache access for concurrent cached `ObjectBinary`
      reads; broader high-level API concurrency is not claimed without
      path-specific tests.
- [x] Keep `inspect` as metadata-only and make `parse` perform a
      fault-tolerant full-object parse summary, then expand `cmd/arksave`
      with local profile/tribe file and directory summaries plus
      save-info/domain JSON exports.
- [x] Keep mutation helpers in an explicit experimental surface that requires an
      output path, including copy, object removal, object-byte upsert, and
      custom-table upsert CLI commands.
- [x] Re-run oracle integration after major parser/API expansion.
