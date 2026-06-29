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
- [x] Split domain models under `arkobject` or subpackages. Equipment item
      construction, stat calculation, default stat tables, and property
      coercion helpers are now separated. Dino construction, colors, lineage,
      traits, and shared object-property coercion helpers are also separated;
      player, tribe, and shared profile property helpers are separated; local
      cluster item, dino, and class-name helpers are separated; inventory
      collection, inventory item, stackable item, and shared reference helpers
      are separated; structure construction, ownership matching, linked
      references, and shared scalar property helpers are separated.
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
- [x] Add `arkapi.NewClusterFromPath` so typed local cluster examples and
      library consumers open local cluster files without direct `arkcluster`
      handling.
- [x] Add `arkapi.ClusterSummaryFromPath` and
      `arkapi.ClusterDirectorySummaryFromPath` so local cluster file and
      directory aggregates are reusable typed results without JSON-only access.
- [x] Surface typed player and tribe exports through domain JSON for CLI and
      library consumers.
- [x] Add `arkapi.NewPlayerFromPath` so examples and library consumers can
      open either save files or local save directories with explicit player or
      tribe fallback behavior.
- [x] Add `arkapi.PlayerInventorySummaryForPlayers` and
      `arkapi.PlayerInventorySummaryFromPath` so player inventory examples use
      reusable typed aggregation instead of local counting logic.
- [x] Add `arkapi.BaseAPI.SummaryForBases` and
      `arkapi.BaseAPI.SummaryWithFaults` so base aggregate counts are reusable
      typed API results instead of caller-local counting logic.
- [x] Add `arkapi.BaseSummaryFromPath` so callers can open local saves and get
      typed base aggregate summaries without manual save lifecycle handling.
- [x] Add `arkapi.BaseComponentStatsFromPath`, then move `base-components`
      example and CLI aggregate workflows onto the typed path helper.
- [x] Add `arkapi.DinoHeatmapSummaryFromPath` and
      `arkapi.StructureHeatmapSummaryFromPath` so callers can get typed
      heatmap summaries without writing JSON output files.
- [x] Add `arkapi.StructureHealthSummaryFromPath`,
      `arkapi.StructureOwnerSummaryFromPath`, and
      `arkapi.StructureTribeOwnershipSummaryFromPath`, then move
      `structure-health`, `structure-owner-count`, `structure-owners`, and
      `structure-owner-locations` CLI aggregate commands plus matching
      structure aggregate examples onto typed structure path helpers while
      preserving parse-fault counts and redaction behavior.
- [x] Add `arkapi.DinoPopulationSummaryFromPath`,
      `arkapi.DinoWildTamableSummaryFromPath`, and
      `arkapi.DinoBabySummaryFromPath` so common dino aggregate examples can
      use typed path helpers without manual save lifecycle handling.
- [x] Move `dinos`, `dino-wild-tamables`, and `dino-babies` CLI aggregate
      commands onto the typed dino path helpers while preserving parse-fault
      counts.
- [x] Add `arkapi.DinoBestStatSummaryFromPath`,
      `arkapi.DinoMostMutatedSummaryFromPath`, and
      `arkapi.DinoWildTamedSummaryFromPath`, then move the matching CLI
      commands and examples onto typed dino path helpers while preserving
      parse-fault reporting.
- [x] Add `arkapi.EquipmentSummaryFromPath`,
      `arkapi.EquipmentSummaryIncludingCryopodSaddlesFromPath`,
      `arkapi.EquipmentSaddleSummaryFromPath`,
      `arkapi.EquipmentBestSummaryFromPath`,
      `arkapi.EquipmentRankStatsFromPath`, and
      `arkapi.EquipmentOwnedSummaryFromPath` so equipment aggregate examples
      can use typed path helpers without manual save lifecycle handling.
- [x] Move `equipment-summary`, `equipment-saddles`,
      `equipment-best`, `equipment-rank`, `equipment-ascendant-weapon-bps`,
      and `equipment-owned-by` CLI aggregate commands and matching aggregate
      examples onto typed equipment path helpers while preserving parse-fault
      counts.
- [x] Add `arkapi.StackableSummaryFromPath` and
      `arkapi.StackableSummaryFromPathWithFaults` plus
      `arkapi.StackableOwnedSummaryFromPath` so stackable aggregate examples
      and CLI commands can use typed path helpers without manual save lifecycle
      handling.
- [x] Add `arkapi.PlayerRosterSummaryFromPath`,
      `arkapi.PlayerAllSummaryFromPath`,
      `arkapi.TribeRosterSummaryFromPath`,
      `arkapi.TribePlayerRelationSummaryFromPath`, and
      `arkapi.PlayerAndTribeDataSummaryFromPath` so player/tribe aggregate
      CLI commands and examples can use typed path helpers without manual
      `PlayerAPI` lifecycle handling.
- [x] Add `arkapi.PlayerUnlockedEngramsFromPath`, then move
      `player_unlocked_engrams` onto the typed path helper while preserving
      sorted unique engram output.
- [x] Add `arkapi.LocalProfileSummaryFromPath`, then move
      `local_profiles` onto the typed path helper while preserving local file
      counts, aggregate metrics, and per-operation warning behavior.
- [x] Add `arkapi.TribeDirectorySummaryFromPath`, then move the `tribes`
      directory CLI path onto the typed helper while preserving aggregate and
      non-redacted row output.
- [x] Add `arkapi.ExportDinoBinaryFromPath` and move
      `dino_export_from_save` onto the typed path helper while preserving the
      explicit output directory contract for structural mutation fixtures.
- [x] Add `arkapi.ExportStructureBinaryFromPath`,
      `arkapi.ExportEquipmentBinaryFromPath`, and
      `arkapi.ExportBaseBinaryFromPath`, then move the remaining structure,
      equipment, and base binary export examples onto typed path helpers.
- [x] Add `arkapi.NewJSONFromPath`, then move save-info export, all-domain
      JSON export, and equipment history snapshots onto shared typed JSON
      path lifecycle handling.
- [x] Move dino and structure heatmap path helpers onto shared typed
      `arkapi.NewDinoFromPath` and `arkapi.NewStructureFromPath` lifecycle
      handling.
- [x] Add `arkapi.StructureAtLocationSummaryFromPath`, then move
      `structure_at_location` onto the typed structure path helper while
      preserving the existing nearby/connected aggregate output.
- [x] Move `structure-heatmap` and `dino-heatmap` CLI commands onto typed
      heatmap export path helpers, including selected-structure heatmap JSON
      export so the structure command keeps its selected-index semantics.
- [x] Add `arkapi.GeneralClassesFromPath` and
      `arkapi.GeneralParseSummaryFromPath`, then move `object_classes` and
      `parse_all` examples onto typed general path helpers while preserving
      staged parse error labels.
- [x] Add path-level typed general object, property-position, class-lookup,
      class-property, and property-filter summaries, then move the matching
      general aggregate examples onto typed path helpers.
- [x] Move `map-summary` CLI/example save-info export onto typed JSON path
      helpers.
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
      ID-table game-object bytes with custom object-name payloads use a shared
      testfixtures helper;
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
      object-reference array fixtures use shared ID-table array writers; and
      player/tribe relation directory fixtures plus dino stats/status object
      byte fixtures are shared by API and CLI tests; stackable API tests and
      benchmarks use shared stackable object fixtures directly; and simple
      ID-table int-property object rows use
      `testfixtures.ObjectBytesWithIntProperty` in save-layer/general tests.
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
