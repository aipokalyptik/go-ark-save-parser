# Project Task Ledger

This is the monitorable task list for the offline-only Go port of upstream
`ark-save-parser`. It exists so progress can be audited from the repository
without reading chat history.

For stable up-front task IDs across the whole goal, see
[`docs/task-inventory.md`](task-inventory.md). This ledger carries the detailed
status notes behind those IDs.

Status markers:

- `[x]`: complete and committed.
- `[~]`: partially implemented or intentionally narrowed; see the note.
- `[ ]`: not complete.
- `[blocked]`: blocked by missing fixture data, upstream behavior, live-server
  validation, or an explicit out-of-scope boundary.

## Current Rollup

| Phase | Status | Done Means | Current Blocking Items |
| --- | --- | --- | --- |
| Phase 1: Oracle Setup | `[x]` Complete for selected-feature parity | Existing Python oracle behavior is reproducible from private data for the chosen offline features. | Expanding Python/oracle coverage for every upstream example is intentionally out of scope. |
| Phase 2: Literal Go Transpilation | `[x]` Closed with documented blockers | Offline Go behavior mirrors runnable upstream behavior closely enough for selected oracle-derived tests/examples to pass or have documented blockers. | Remaining items are fixture-gated, upstream-blocked, intentionally outside Python-oracle expansion, or live-server-unverified mutation work. |
| Phase 3: Idiomatic Go Refactor | `[x]` Complete | Translated behavior is organized into stable Go packages and CLI surfaces without losing oracle parity. | Phase 3 is closed; keep future changes in Phase 4 unless a verified regression reopens Phase 3. |
| Phase 4: Documentation And Production Readiness | `[~]` Closed except blocked oracle rerun | Another engineer can build, test, run, and extend the project without Python/private context. | Go verification, provided-data E2E, CLI smoke, and final review are complete; the selected oracle comparison full-suite rerun is blocked by upstream runtime cost in the supplied private save. |

## Execution Mode

Work is phase-gated from this point forward:

- Phase 1 is closed for the selected offline parity scope. Do not add Python
  oracle examples, improve upstream Python code, or expand oracle coverage
  unless a Phase 2 blocker cannot be understood without a narrowly scoped
  existing-oracle check.
- Phase 2 is closed for the selected offline, fixture-backed scope. Do not
  reopen Phase 2 for Python oracle expansion; only reopen it when a new Go
  failing test or provided-data failure exposes a concrete offline parser/API
  parity defect.
- Phase 3 is closed after the typed API, fixture, package, performance, CLI, and regression rows were completed.
- Phase 4 is closed for documentation, release-build, smoke, provided-data, and production-readiness review work. The selected oracle comparison full-suite rerun remains blocked by upstream runtime cost in the supplied private save; do not expand Python oracle coverage to work around it.
- Status docs may still be updated in the same commit as implementation work so
  the repository remains monitorable.

## Operating Rules

- [x] Keep FTP and RCON out of scope.
- [x] Keep cluster support local-file-only.
- [x] Treat mutation APIs as explicit copied-save workflows that are
      experimental and live-server-unverified.
- [x] Keep private saves, oracle raw output, snapshots, debug dumps, private
      manifests, and extracted save data out of git.
- [x] Push coherent task-group commits to `main`.
- [x] Use sanitized docs for progress reporting; detailed private oracle values
      stay under `.oracle/output`.
- [x] Do not expand the Python codebase or oracle suite beyond selected-feature
      parity needs; new work should target Go offline feature behavior.
- [x] Prefer expanding Go code coverage, examples, and read-only provided-data
      E2E tests for chosen offline features.
- [x] Update this ledger in the same commit whenever a new task, blocker, or
      completion status is discovered.

## Phase 1: Oracle Setup

Done when Python oracle behavior is reproducible from private backup data for
the selected offline parity features and sanitized status exists for the
reference oracle work that was already run.

- [x] Create repository hygiene:
  - [x] Initialize `go.mod`.
  - [x] Add `.gitignore` entries for `.oracle/`, extracted saves, temp DBs,
        snapshots, debug dumps, generated private output, and local env files.
  - [x] Add MIT license attribution compatible with upstream.
  - [x] Create public GitHub repository `go-ark-save-parser`.
  - [x] Push `main`.
- [x] Prepare oracle workspace:
  - [x] Verify `~/Downloads/SavedArks.tar.bz2` exists.
  - [x] Extract private backup under `.oracle/data`.
  - [x] Clone upstream at commit
        `4f7cc91fb96a080321bfbc884ba81bd897f72c49` under
        `.oracle/upstream`.
  - [x] Install/use Python 3.13+ for upstream dependency compatibility.
  - [x] Create `.oracle/venv`.
  - [x] Install upstream editable package and test tooling.
- [x] Inventory private backup:
  - [x] Detect `.ark` files.
  - [x] Detect `.arkprofile` files.
  - [x] Detect `.arktribe` files.
  - [x] Detect local cluster-like extensionless files.
  - [x] Detect local compact tribute index files.
  - [x] Generate private `.oracle/manifest.json`.
  - [x] Commit sanitized count-only summary.
- [x] Run upstream oracle tests:
  - [x] Run upstream packaged `pytest`.
  - [x] Record packaged-test blockage on missing upstream `tests/test_data`.
  - [x] Run upstream `testbench/pytest` against usable `.ark` saves.
  - [x] Record sanitized pass/fail/skip status.
- [x] Run selected offline examples:
  - [x] Classify read-only, mutation-copy, export-producing,
        heatmap-producing, local-cluster, network-only skip, and impossible
        examples in `docs/upstream-oracle-classification.md`.
  - [x] Add privacy-safe oracle comparison harness in
        `scripts/oracle_compare.py`.
  - [x] Capture private oracle comparison output under `.oracle/output`.
  - [x] Commit sanitized aggregate status in
        `docs/oracle-comparison-summary.md`.
  - [x] Read-only selected examples are covered by Go counterparts and
        sanitized aggregate comparisons where existing oracle evidence is
        useful.
  - [x] Mutation examples are structurally represented by copy/write/reopen
        helpers where feasible; live-server behavior is unverified by design.
- [x] Phase review:
  - [x] Review oracle completeness and privacy boundaries.
  - [x] Commit sanitized Phase 1 report.

## Phase 2: Literal Go Transpilation

Done when the Go code mirrors upstream offline behavior closely enough for
oracle-derived tests/examples to pass, or for blockers to be explicitly
documented.

### Core Binary Layer

- [x] Port byte reader/writer behavior.
- [x] Port bounds, seek, skip, peek, and remaining-byte behavior.
- [x] Port little-endian integers, floats, doubles, and bools.
- [x] Port ARK strings, including positive null-terminated byte strings and
      negative UTF-16LE strings.
- [x] Port UUID byte-order semantics.
- [x] Port name-table lookup and missing-name behavior.
- [x] Port zlib inflation.
- [x] Port ARK wildcard decompression.
- [x] Add primitive unit tests.

### Save Access

- [x] Port SQLite `.ark` loading using pure-Go SQLite.
- [x] Port save context/header/name-table/custom-table reads.
- [x] Port game-table object binary access.
- [x] Port actor transform custom table parsing.
- [x] Port class lookup and object enumeration.
- [x] Add synthetic SQLite tests.
- [x] Add gated private oracle tests.

### Property Parser

- [x] Port terminator handling.
- [x] Port primitive properties.
- [x] Port raw byte and enum `ByteProperty` encodings.
- [x] Port object references.
- [x] Port `SoftObjectProperty`, `NameProperty`, `Int8Property`,
      `Int16Property`, `Int64Property`, `UInt16Property`, and inline
      `UInt64Property`.
- [x] Port generic structs and struct arrays.
- [x] Port simple arrays, maps, and sets.
- [x] Port generic struct-valued maps.
- [x] Port enum-keyed map descriptor payloads.
- [x] Add body-skip alignment for unsupported map/set/struct-key encodings.
- [x] Preserve raw fallback for unknown structs and unknown property types.
- [x] Port packed `Vector`, `Rotator`, `Quat`, `Color`, and `LinearColor`.
- [x] Add declared-size realignment and recoverable overread handling.
- [x] Preserve partial struct containers on recoverable profile overread.
- [x] Continue after aligned malformed compound properties in partial parsing
      while still returning the recovery error.
- [blocked] Port remaining compound payload encodings discovered by future
      oracle failures; currently no open Phase 2 failure exposes a concrete
      unported encoding.
- [blocked] Port legacy property/object parsing where a runnable local oracle
      path exists. Upstream uses separate legacy binary/property parsers for
      pre-Unreal-5.5 archives and legacy cryopod payloads; the Go port reports
      typed unsupported legacy errors until a fixture proves exact behavior.

### Object Model

- [x] Port generic game object headers and property containers.
- [x] Add inventory and inventory-item wrappers.
- [x] Add object-owner wrappers.
- [x] Add structure wrappers and inventory metadata.
- [x] Add equipment wrappers and crafter metadata.
- [x] Add player and tribe wrappers for profile/archive containers.
- [x] Add dino wrappers for identity, status, ownership, stats, colors,
      mutations, baby stage, ancestry, and traits.
- [x] Add base summary wrappers for grouped structures.
- [x] Preserve raw binary/property positions and encoded byte spans for
      structural mutation tests.
- [x] Complete read-first wrappers required by the chosen offline examples for
      inventory, owner, trait, dino, structure, equipment, stackable, player,
      tribe, base, crafter, and local cluster workflows; add future fields only
      when a concrete Go failure or fixture exposes the need.

### Offline APIs

- [x] Port General API object queries and fault collection.
- [x] Port local profile and tribe discovery/loading.
- [x] Port save-contained player and tribe parsing, including game-table
      objects and embedded `GameModeCustomBytes`.
- [x] Port player lookups, tribe relationships, player aggregates,
      unlocked-engrams aggregation, pawn lookup, inventory lookup, and location
      lookup.
- [x] Port local tribute discovery/loading, fault-preserving directory reads,
      and JSON/CLI directory fault reporting for malformed local tribute index
      files, including direct `tribute` directory summaries.
- [x] Port local cluster discovery/loading, uploaded item/dino summaries, typed
      projections, parse-status reporting, and JSON export.
- [x] Port Structure API class, owner, location, connected-structure, heatmap,
      inventory-container, and selected-property fast scan paths.
- [x] Port Stackable API counts, categories, owner filtering, and JSON export.
- [x] Port Equipment API weapon/armor/saddle/shield queries, owner filtering,
      crafted state, quality/rating/durability filters, stat calculations,
      ranking helpers, inventory-state summaries, and modern cryopod saddle
      extraction.
- [x] Port Dino API class/tamed/wild/baby/stat/mutation/trait/ownership/
      ancestry/pedigree-base/heatmap/filter helpers and modern cryopod
      dino/status extraction.
- [x] Port Base API grouping, point lookup, minimum count filtering, all-base
      discovery, and fast base-component statistics.
- [x] Port save-info and implemented domain JSON exports.
- [x] Add explicit-output `export_all_items` example and manifest.
- [x] Add local multi-save `equipment_history` example.
- [blocked] Finish full dino edge behavior:
  - [blocked] Legacy/modded cryopod variants require a concrete supported
        fixture or runnable oracle case.
  - [x] Fixture-backed cryopod-location fallback uses the containing cryopod
        item's `ActorTransforms` entry for modern embedded cryopodded dinos
        while preserving `InCryopod` location semantics.
  - [x] Typed cryopod payload errors classify malformed embedded dino payload
        parse failures while preserving wrapped parser errors for `errors.Is`.
  - [x] Typed cryopod payload errors classify unsupported embedded saddle
        versions so fault-tolerant callers can distinguish legacy/modded
        saddle payload failures from generic parse failures.
  - [x] Typed pedigree tree helpers and domain JSON pedigree trees beyond flat
        child/descendant UUID references.
  - [blocked] Full upstream/private `dino_pedigrees` oracle comparison remains
        blocked by malformed cryopod parsing before stable aggregate output.
- [blocked] Finish full structure/base edge behavior:
  - [x] Add exact full-parse owner/location grouping through
        `StructureAPI.OwnerLocationsFullWithFaults` for fixture-sized parity
        checks while preserving selected-property `structure_owner_locations`
        as the large-save fast path.
  - [x] `structure_owner_locations` reports skipped candidate counts for
        structures without usable owner or location data while preserving valid
        selected-property owner/location buckets.
  - [blocked] `structure_heatmap` oracle comparison, blocked because upstream
        indexes out-of-range cells on the supplied private save.
  - [x] Base export/import read/write parity where local-copy structural tests
        are feasible, including generated export-to-import row checks.
- [blocked] Finish full equipment edge behavior:
  - [x] Move high-rating equipment ranking candidate selection and aggregate
        stats into typed `arkapi` helpers used by the example.
  - [x] Add upstream family/slot default stat tables for equipment durability,
        armor, hypothermal resistance, and hyperthermal resistance.
  - [x] Add a Go regression guard that every curated upstream equipment
        blueprint list classifies to the expected kind.
  - [x] Cover cursed shield durability and shield armor defaults in Go stat
        table tests.
  - [x] Cover inferable cursed weapon durability defaults in Go stat table
        tests.
  - [x] Cover base and cursed Tek sword durability defaults in Go stat table
        tests.
  - [x] Compute equipment average-stat ranking values from upstream-style
        effective internal stats, including default durability, damage, armor,
        and zero-default insulation behavior when `ItemStatValues` entries are
        absent or below upstream minimums.
  - [x] Skip unsupported embedded cryopod saddle payloads in the plain read API
        while reporting them through the fault-collecting API.
  - [x] Classify unsupported embedded cryopod saddle payload versions with a
        typed `arkobject.CryopodPayloadError` surfaced through fault-collecting
        APIs.
  - [x] Model generic equipment `CustomItemDatas` presence/count metadata,
        include it in equipment summaries, JSON export rows, and the
        `equipment_summary` example output.
  - [blocked] Exact equipment ranking count parity and full private comparison
        for average-stat aggregates would require expanding the Python oracle
        suite; Go fixtures cover `Ranked` and `BestAverageStat`.
  - [blocked] Legacy/modded cryopod saddle payloads and cosmetics require
        concrete fixtures or failures.
  - [blocked] Remaining long-tail default armor/stat table parity should resume
        only when new concrete mismatches are found.
- [blocked] Finish richer local cluster item/dino domain models:
  - [x] Model uploaded item type with typed constants and helper methods while
        preserving string-based filters and JSON output.
  - [x] Add enum-based uploaded item filters, version/parse helper methods, and
        embedded uploaded-dino component class-name summaries.
  - [x] Add explicit uploaded-dino parse status helpers and counts for parsed,
        unsupported-version, parse-error, and unparsed local cluster uploads.
  - [x] Add typed uploaded-item aggregate summaries for item type counts,
        version support counts, crafted item counts, total quantity, and max
        rating/quality.
  - [x] Add typed uploaded-dino aggregate summaries for parsed/error counts,
        version support counts, component presence, and embedded object totals.
  - [x] Include uploaded item version support and uploaded dino parse status,
        version support, parsed-archive state, and component class summaries in
        cluster JSON exports.
  - [x] Include uploaded item short names and uploaded dino primary/short class
        names in typed local cluster models and JSON exports.
  - [x] Include embedded uploaded-dino ID, tamed/name/sex/baby/dead flags, and
        stat level summaries in typed local cluster models and JSON exports
        when the local upload archive contains parseable dino and status
        objects.
  - [x] Include embedded uploaded-dino ID, tamed metadata, sex/baby/dead flags,
        and stat level summaries in local cluster CLI detail output.
  - [x] Include embedded uploaded-dino identity, tamed/sex/baby/dead, and stat
        aggregate counts in typed local cluster summaries and `cluster-summary`
        CLI output.
  - [x] Include embedded uploaded-dino base/current level totals, averages, and
        maxima in typed local cluster summaries, `cluster-summary` CLI output,
        and JSON-derived directory summary aggregation.
  - [x] Include embedded uploaded-dino identity/stat aggregate counters in the
        idiomatic `cluster_typed` Go example.
  - [x] Include uploaded item short names and uploaded dino primary/short class
        names in local cluster CLI summaries.
  - [x] Include uploaded-dino parse-status aggregate counts in local cluster
        CLI summaries.
  - [x] Include explicit unparsed uploaded-dino counts in the typed local
        cluster example output.
  - [x] Add directory-level local cluster aggregate summaries for total files,
        objects, items, dinos, parse errors, uploaded-item summaries, and
        uploaded-dino summaries, and include them in directory JSON exports and
        CLI directory summaries.
  - [x] Add fault-preserving local cluster directory reads so valid cluster
        uploads survive alongside broken local cluster files.
  - [x] Expose local cluster directory file faults in JSON export and CLI
        `cluster`, `cluster-summary`, and `export-cluster-json` directory
        output.
  - [blocked] Add richer item/dino fields only when new local-file fixtures
        expose them.
- [blocked] Finish remaining Player/Tribe edge behavior not covered by parsed local
      archives, game-table objects, or embedded `GameModeCustomBytes`:
  - [x] Move save-contained player pawn inventory indexing and upstream-style
        inventory item counting from the example into typed `arkapi` helpers.
  - [x] Move player inventory aggregate summary calculations into typed
        `arkapi.PlayerInventorySummaryForPlayers` and
        `arkapi.PlayerInventorySummaryFromPath`.
  - [x] Move the `player_all` example aggregate into typed
        `arkapi.PlayerAllSummary`.
  - [x] Extend `arkapi.PlayerAndTribeDataSummary` with typed relationship edge
        counters for tribes with inactive members and tribes without active
        players.
  - [x] Exercise typed local player death, level, and experience average
        helpers through the `local_profiles` example smoke test.
  - [x] Include average experience in typed player directory summaries and the
        `players <directory>` CLI aggregate output.
  - [x] Include highest level and highest experience in typed player directory
        summaries and the `players <directory>` CLI aggregate output.
  - [x] Include total and average member counts in typed tribe directory
        summaries and the `tribes <directory>` CLI aggregate output.
  - [x] Add fault-preserving local `.arkprofile` and `.arktribe` batch reads so
        valid local player/tribe records survive alongside broken archive files.
  - [x] Add fault-returning player roster, player-all, and tribe roster summary
        helpers so local batch summaries can report partial data explicitly.
  - [blocked] Continue porting remaining upstream player/tribe edge cases only
        as chosen offline examples or Go failures expose them.
- [blocked] FTP and RCON modules are intentionally omitted.

### Mutation APIs

- [x] Port copy-based save modification helper.
- [x] Port object removal by UUID.
- [x] Port object removal by class substring.
- [x] Port object-byte upsert on copied saves.
- [x] Port custom-table byte upsert on copied saves.
- [x] Require explicit output paths.
- [x] Add structural write/reopen/reparse tests.
- [x] Document live-server validation as out of scope.
- [blocked] Upstream generated-blueprint insertion is classified as mutation-copy
      only; live-server acceptance remains unverified.
- [blocked] Upstream base import/customize examples have partial structural coverage
      through exported raw structure rows and `ImportBaseBinary` reinsert into
      explicit copied saves. Moving structures, inventory expansion,
      customization, owner replacement, and live-server acceptance remain
      unverified.
  - [x] Generated base binary exports round-trip through `ImportBaseBinary`
        into reopenable copied saves with byte-for-byte row checks.
- [blocked] Upstream structure modification examples have partial structural
      coverage through standalone `structure_export_from_save` raw structure
      rows and `ImportStructureBinary` reinsert into explicit copied saves.
      Health/owner mutation and live-server acceptance remain unverified.
  - [x] Generated structure binary exports round-trip through
        `ImportStructureBinary` into reopenable copied saves.
- [blocked] Upstream dino extract/reinsert examples have partial structural coverage
      through exported direct-save dino, status, and inventory rows plus
      `ImportDinoBinary` reinsert into explicit copied saves. Cryopod insertion
      into target inventories and generated location changes remain
      unverified.
  - [x] Generated dino binary exports round-trip through `ImportDinoBinary`
        into reopenable copied saves.
- [blocked] Upstream dino trait/stat/growth mutation examples have partial
      structural coverage through `ReplaceObjectPropertyBinary`, which can
      replace a parsed object's full encoded property record by name and
      position in an explicit copied save, then reopen and reparse it.
      Semantic trait/stat/growth authoring and live-server acceptance remain
      unverified.
- [blocked] Upstream generated-blueprint/equipment insertion example has partial
      structural coverage through exported equipment rows and
      `ImportEquipmentBinary` reinsert into explicit copied saves. Generated
      blueprint construction, insertion into target inventories, and
      live-server acceptance remain unverified.
  - [x] Generated equipment binary exports round-trip through
        `ImportEquipmentBinary` into reopenable copied saves.
- [x] Add structural mutation tests for upstream dino trait/stat/growth
      examples where local-copy behavior can be represented without claiming
      live-server safety.

### Examples And Oracle Comparisons

- [x] `map_summary`.
- [blocked] `parse_all`: implemented and smoke-tested; private comparison is
      manual because full save parsing is runtime-heavy and broad Python oracle
      expansion is out of scope.
- [x] `object_classes`.
- [x] `object_summary`.
- [x] `property_positions`.
- [x] `class_lookup`.
- [x] `class_property_summary`.
- [x] `property_filter`.
- [x] `local_profiles`.
- [x] `player_all`.
- [x] `player_list`.
- [x] `tribe_list`.
- [x] `player_and_tribe_data`.
- [x] `player_tribe_links`.
- [x] `player_inventory`.
- [x] `player_inventories`.
- [x] `player_unlocked_engrams`.
- [x] `dino_filter`.
- [x] `dino_best_stat`.
- [x] `dino_best_base_stat`.
- [x] `dino_most_mutated`.
- [x] `dino_babies`.
- [x] `dino_wild_tamables`.
- [x] `dino_wild_tamed`.
- [x] `dino_export_from_save`: represented as an explicit-output structural
      export that writes copied direct-save dino rows plus linked status and
      inventory rows when present. Cryopod insertion and live-server validation
      remain mutation-copy-adjacent and unverified.
- [x] `dino_heatmap`.
- [blocked] `dino_cryopod_location` oracle comparison: Go now has
      fixture-backed cryopod item transform fallback for modern embedded
      cryopodded dinos, but upstream/private malformed cryopod paths still
      block stable oracle output.
- [blocked] `dino_pedigrees`: upstream/private malformed cryopod path blocks
      stable oracle output; Go now exposes typed pedigree trees and nested
      domain JSON pedigree branches for parsed tamed dinos.
- [x] `stackable_count`.
- [x] `stackable_owned_by`.
- [x] `domain_json_stackables`.
- [x] `equipment_best`.
- [x] `equipment_summary`.
- [x] `equipment_rank` stable fields.
- [x] `equipment_ascendant_weapon_bps`.
- [x] `equipment_saddles` direct saddle fields.
- [x] `equipment_owned_by`.
- [x] `equipment_export_from_save`: represented as an explicit-output
      structural export that writes copied equipment item rows. Generated
      blueprint construction, inventory insertion, and live-server validation
      remain mutation-copy-adjacent and unverified.
- [x] `domain_json_equipment` longneck aggregate.
- [x] `structure_owner_count`.
- [x] `structure_owners`.
- [x] `structure_owner_locations` stable multi-structure cells.
- [x] `structure_at_location`.
- [blocked] `structure_heatmap` oracle comparison: upstream out-of-range
      heatmap indexing on supplied private save.
- [x] `structure_export_from_save`: represented as an explicit-output
      structural export that writes copied raw structure rows and structure
      location JSON. Health/owner mutation, binary editing, and live-server
      validation remain mutation-copy-adjacent and unverified.
  - [x] Generated structure exports are covered by export-to-import copied-save
        round-trip tests.
- [x] `base_components`.
- [x] `base_export_from_save`: represented as an explicit-output structural
      export that writes base metadata, copied raw structure rows, and
      structure location JSON. Inventory item expansion, binary import, and
      live-server validation remain mutation-copy-adjacent and unverified.
  - [x] Generated base exports are covered by export-to-import copied-save
        round-trip tests.
- [x] `cluster_json`.
- [x] `cluster_typed`.
- [x] `local_tribute`.
- [x] `tribute_json`.
- [x] `export_json`.
- [blocked] `export_all_items`: implemented and smoke-tested; default oracle
      comparison is deferred because full export is too slow on the large
      private save and broad Python oracle expansion is out of scope.
- [blocked] `equipment_history` oracle comparison: supplied backup lacks a
      timestamped historical save sequence.
- [x] `logging_config`.
- [x] `mutation_copy`.

## Phase 3: Idiomatic Go Refactor

Done when the translated implementation is organized into stable Go packages,
CLI tools, and reusable APIs without losing oracle parity.

- [x] Package split:
  - [x] `arkbinary`.
  - [x] `arksave`.
  - [x] `arkproperty`.
  - [x] `arkobject`.
  - [x] `arkprofile`.
  - [x] `arkapi`.
  - [x] `arkcluster`.
  - [x] `arktribute`.
  - [x] `arkmutation`.
  - [x] `cmd/arksave`.
- [x] API cleanup:
  - [x] Add typed APIs for implemented player, tribe, dino, structure,
        equipment, stackable, base, cluster, tribute, and JSON workflows.
  - [x] Add typed local cluster uploaded-item type constants and helper methods.
  - [x] Use explicit errors and fault collections for partial object parses.
  - [x] Require explicit output paths for mutation operations.
  - [x] Split or further organize large domain models under `arkobject` or
        subpackages once behavior stabilizes; equipment item construction,
        stat calculation, default stat tables, and property coercion helpers
        are now split into focused files; dino construction, colors, lineage,
        traits, and shared object-property coercion helpers are also split with
        package-shape regression tests; player, tribe, and shared profile
        property helpers are split with package-shape regression coverage;
        cluster item, dino, and class-name helpers are split with
        package-shape regression coverage; inventory collection, inventory
        item, stackable item, and shared reference helpers are split with
        package-shape regression coverage; structure construction, ownership
        matching, linked references, and shared scalar property helpers are
        split with package-shape regression coverage.
  - [x] Replace remaining Python-shaped compatibility helpers where typed Go
        surfaces now exist:
    - [x] Move `object_summary` and `class_property_summary` direct save
          parsing behind typed `arkapi.GeneralAPI` summary helpers.
    - [x] Move `class_lookup` selected-property class counting behind a typed
          `arkapi.GeneralAPI` summary helper.
    - [x] Move `property_filter` and `property_positions` aggregate counting
          behind typed `arkapi.GeneralAPI` summary helpers.
    - [x] Move `parse_all` and `object_classes` save-level counting/listing
          behind typed `arkapi.GeneralAPI` helpers.
    - [x] Move repeated GeneralAPI example save-open/close handling behind
          typed general path helpers.
    - [x] Add `arkapi.GeneralClassesFromPath` and
          `arkapi.GeneralParseSummaryFromPath`, then move `object_classes`
          and `parse_all` example save-open/close handling onto typed path
          helpers while preserving staged parse error labels.
    - [x] Add path-level typed object, property-position, class-lookup,
          class-property, and property-filter general summaries, then move the
          matching general aggregate examples onto typed path helpers.
    - [x] Move `map_summary` save opening and save-info export behind typed
          `arkapi.ExportSaveInfoFromPath`.
    - [x] Move repeated player/tribe save-or-directory fallback logic behind
          typed `arkapi.NewPlayerFromPath`.
    - [x] Move `local_tribute` file/directory aggregate counting behind typed
          `arkapi` tribute summary helpers.
    - [x] Move `tribute_json` file/directory JSON selection behind typed
          `arkapi.ExportTributePathJSON` so examples reuse fault-preserving
          directory handling.
    - [x] Move `cluster_json` file/directory JSON selection behind typed
          `arkapi.ExportClusterPathJSON` so examples reuse fault-preserving
          directory handling.
    - [x] Move `cluster_typed` local cluster file opening behind typed
          `arkapi.NewClusterFromPath`.
    - [x] Add typed local cluster file and directory aggregate helpers through
          `arkapi.ClusterSummaryFromPath` and
          `arkapi.ClusterDirectorySummaryFromPath`.
    - [x] Move `cluster`, `cluster-summary`, and `export-cluster-json` CLI
          file/directory lifecycle handling onto typed cluster path helpers,
          and expose crafted upload status in cluster JSON item models so the
          CLI does not need direct `arkcluster` access for typed summaries.
    - [x] Move `tribute` and `export-tribute-json` CLI file/directory
          lifecycle handling onto typed tribute path helpers while preserving
          directory fault reporting, redaction, and explicit output-path
          behavior.
    - [x] Move `player_all` aggregate counting behind typed
          `arkapi.PlayerAllSummary`.
    - [x] Move `player_inventories` aggregation behind typed
          `arkapi.PlayerInventorySummaryFromPath`.
    - [x] Move `player_inventory` per-player inventory/location lookup behind
          typed `arkapi.PlayerInventoryLookupFromPath`.
    - [x] Add `arkapi.PlayerUnlockedEngramsFromPath`, then move
          `player_unlocked_engrams` sorted unique engram output onto the typed
          player path helper.
    - [x] Add `arkapi.LocalProfileSummaryFromPath`, then move
          `local_profiles` local file counts, aggregate metrics, and
          per-operation warnings onto the typed player path helper.
    - [x] Add typed player profile and tribe file summary helpers, then move
          the `players <file.arkprofile>` and `tribes <file.arktribe>` CLI
          paths off direct `arkprofile` lifecycle handling while preserving
          archive metadata output before parse errors.
    - [x] Add `arkapi.TribeDirectorySummaryFromPath`, then move the `tribes`
          directory CLI aggregate and non-redacted row output onto the typed
          player path helper.
    - [x] Add `arkapi.PlayerDirectorySummaryFromPath`, then move the `players`
          directory CLI aggregate, redaction, and non-redacted row output onto
          the typed player path helper.
    - [x] Move equipment max-damage example aggregation onto existing typed
          `arkapi.EquipmentAPI.BestWeaponDamage`.
    - [x] Move `equipment_best` filter-plus-best selection onto typed
          `arkapi.EquipmentAPI` fault-collecting best-item helpers.
    - [x] Add `arkapi.EquipmentBestSummaryFromPath` and move
          `equipment-best` CLI/example output onto the typed path helper.
    - [x] Add `arkapi.EquipmentRankStatsFromPath` and move `equipment-rank`
          CLI/example output onto the typed path helper.
    - [x] Move `equipment_ascendant_weapon_bps` count/max-damage aggregation
          onto existing typed `arkapi.EquipmentAPI.SummaryWithFaults`.
    - [x] Move the `equipment_ascendant_weapon_bps` example save-open/close
          lifecycle onto `arkapi.EquipmentSummaryFromPath`.
    - [x] Move `equipment_history` snapshot identity, diff logic, manifest
          reading, and report assembly onto typed `arkapi` equipment history
          helpers.
    - [x] Move `equipment_owned_by` owner filtering and max damage aggregation
          onto typed `arkapi.EquipmentAPI.OwnedSummaryWithFaults`.
    - [x] Move saddle max-armor and wild-tamed max-level example aggregation
          onto typed `arkapi.EquipmentAPI.BestArmor` and
          `arkapi.DinoAPI.MaxCurrentLevel`.
    - [x] Move `equipment_saddles` direct/cryopod saddle counting and max armor
          aggregation onto typed `arkapi.EquipmentAPI.SaddleSummaryWithFaults`.
    - [x] Move `equipment_summary` canonical kind/blueprint counting onto
          typed `arkapi.EquipmentAPI.SummaryIncludingCryopodSaddlesWithFaults`.
    - [x] Move canonical equipment blueprint-list composition onto
          `arkapi.CanonicalEquipmentBlueprints`.
    - [x] Move repeated read-only EquipmentAPI example save-open/close handling
          behind typed `arkapi.NewEquipmentFromPath`.
    - [x] Move `dino_babies` wild/tamed counting onto typed
          `arkapi.DinoAPI.BabySummaryWithFaults`.
    - [x] Move `dino_filter` total/tamed/wild/cryopodded/class counting onto
          typed `arkapi.DinoAPI.PopulationSummaryWithFaults`.
    - [x] Move repeated read-only DinoAPI example save-open/close handling
          behind typed `arkapi.NewDinoFromPath`.
    - [x] Move dino stat CLI token parsing/formatting for `dino_best_stat` and
          `dino_best_base_stat` onto typed `arkobject.DinoStat` helpers.
    - [x] Move `dino_heatmap` cryopod filtering, fault handling, and summary
          generation onto typed `arkapi.DinoAPI.HeatmapSummaryWithFaults`.
    - [x] Move `dino_heatmap` save opening, summary JSON encoding, and
          explicit output writing behind typed
          `arkapi.ExportDinoHeatmapSummaryJSONFromPath`.
    - [x] Add save-info and domain JSON path helpers, including redacted
          variants, then move `export-json` and `export-domain-json` CLI
          workflows off direct save lifecycle handling.
    - [x] Add path-level typed dino heatmap summaries through
          `arkapi.DinoHeatmapSummaryFromPath`.
    - [x] Move `structure_heatmap` structure loading, fault handling, and
          summary generation onto typed
          `arkapi.StructureAPI.HeatmapSummaryWithFaults`.
    - [x] Move `structure_heatmap` save opening, summary JSON encoding, and
          explicit output writing behind typed
          `arkapi.ExportStructureHeatmapSummaryJSONFromPath`.
    - [x] Add path-level typed structure heatmap summaries through
          `arkapi.StructureHeatmapSummaryFromPath`.
    - [x] Add path-level typed structure health, owner, and tribe-ownership
          summaries through `arkapi.StructureHealthSummaryFromPath`,
          `arkapi.StructureOwnerSummaryFromPath`, and
          `arkapi.StructureTribeOwnershipSummaryFromPath`, then move
          `structure-health`, `structure-owner-count`, `structure-owners`, and
          `structure-owner-locations` CLI aggregate commands onto typed
          structure path helpers.
    - [x] Add path-level typed dino population, wild tamable, and baby
          summaries through `arkapi.DinoPopulationSummaryFromPath`,
          `arkapi.DinoWildTamableSummaryFromPath`, and
          `arkapi.DinoBabySummaryFromPath`.
    - [x] Move `dinos`, `dino-wild-tamables`, and `dino-babies` CLI
          aggregate commands onto typed dino path helpers while preserving
          parse-fault counts.
    - [x] Add path-level typed dino best-stat, most-mutated, and wild-tamed
          summaries through `arkapi.DinoBestStatSummaryFromPath`,
          `arkapi.DinoMostMutatedSummaryFromPath`, and
          `arkapi.DinoWildTamedSummaryFromPath`, then move matching CLI
          commands and examples onto typed dino path helpers while preserving
          parse-fault reporting.
    - [x] Add path-level typed equipment aggregate, cryopod-saddle aggregate,
          saddle, and owned-by summaries through
          `arkapi.EquipmentSummaryFromPath`,
          `arkapi.EquipmentSummaryIncludingCryopodSaddlesFromPath`,
          `arkapi.EquipmentSaddleSummaryFromPath`, and
          `arkapi.EquipmentOwnedSummaryFromPath`.
    - [x] Move `equipment-summary`, `equipment-saddles`,
          `equipment-ascendant-weapon-bps`, and `equipment-owned-by` CLI
          aggregate commands onto typed equipment path helpers while preserving
          parse-fault counts.
    - [x] Add path-level typed stackable aggregate, fault-preserving
          aggregate, and owned-by summaries through
          `arkapi.StackableSummaryFromPath`,
          `arkapi.StackableSummaryFromPathWithFaults`, and
          `arkapi.StackableOwnedSummaryFromPath`.
    - [x] Add path-level typed player/tribe aggregate summaries through
          `arkapi.PlayerRosterSummaryFromPath`,
          `arkapi.PlayerAllSummaryFromPath`,
          `arkapi.TribeRosterSummaryFromPath`,
          `arkapi.TribePlayerRelationSummaryFromPath`, and
          `arkapi.PlayerAndTribeDataSummaryFromPath`, then move the roster and
          relation CLI commands plus matching examples onto those helpers.
    - [x] Move repeated pure StructureAPI example save-open/close handling
          behind typed `arkapi.NewStructureFromPath`.
    - [x] Move `structure_health`, `structure_owner_count`, and
          `structure_owners` example save-open/close handling onto typed
          structure path summary helpers.
    - [x] Add `arkapi.StructureAtLocationSummaryFromPath`, then move
          `structure_at_location` nearby/connected counts onto the typed
          structure path helper.
    - [x] Move `structure_owner_locations` mixed structure/player save access
          behind typed `arkapi.StructureOwnerLocationsFromPathWithFaults`.
    - [x] Move repeated read-only StackableAPI example save-open/close handling
          behind typed `arkapi.NewStackableFromPath`.
    - [x] Move `base_components` save-open/close handling behind typed
          `arkapi.NewBaseFromPath`.
    - [x] Add reusable typed base aggregate summaries through
          `arkapi.BaseAPI.SummaryForBases` and
          `arkapi.BaseAPI.SummaryWithFaults`.
    - [x] Add path-level typed base aggregate summaries through
          `arkapi.BaseSummaryFromPath`.
    - [x] Add path-level typed base component summaries through
          `arkapi.BaseComponentStatsFromPath`, then move the
          `base-components` example and CLI aggregate workflow onto the typed
          path helper.
    - [x] Add path-level typed dino binary export through
          `arkapi.ExportDinoBinaryFromPath`, then move
          `dino_export_from_save` onto the helper while preserving explicit
          output directory handling.
    - [x] Add path-level typed structure, equipment, and base binary exports
          through `arkapi.ExportStructureBinaryFromPath`,
          `arkapi.ExportEquipmentBinaryFromPath`, and
          `arkapi.ExportBaseBinaryFromPath`, then move their examples onto the
          helpers while preserving explicit output directory handling.
    - [x] Add `arkapi.NewJSONFromPath`, then move save-info export,
          all-domain JSON export, and equipment history snapshots onto shared
          typed JSON path lifecycle handling.
    - [x] Move dino and structure heatmap path helpers onto shared typed
          `arkapi.NewDinoFromPath` and `arkapi.NewStructureFromPath`
          lifecycle handling.
    - [x] Move `structure-heatmap` and `dino-heatmap` CLI commands onto typed
          heatmap export path helpers, including selected-structure heatmap
          JSON export so the structure command keeps selected-index semantics.
    - [x] Add `arkapi.GeneralAPI.SaveInfo` and keep general CLI commands on
          typed `arkapi.GeneralAPI`/JSON summary helpers; dedicated path
          helpers cover save-info export plus the implemented general
          aggregate examples.
    - [x] Move `export_all_items` domain export loop, manifest writing, and
          save-open/close handling onto typed
          `arkapi.ExportAllDomainsFromPath`.
    - [x] Keep remaining low-level example access limited to generic
          inspection and mutation-copy workflows where direct save/file
          handling is the feature.
- [x] Performance pass:
  - [x] Add benchmarks for full save open/object enumeration, object parse,
        query filters, and JSON export.
  - [x] Add selected-property scans for expensive structure/base workflows.
  - [x] Add opt-in object row cache controls on `arksave.Save`, plus cached
        object-parse benchmark coverage.
  - [x] Add safe concurrency only where tests prove no behavior drift; currently
        covered for opt-in object row cache reads only, with broader high-level
        API concurrency intentionally unclaimed.
- [x] CLI:
  - [x] `inspect`.
  - [x] `parse`.
  - [x] `map-summary`.
  - [x] `object-classes`.
  - [x] `object-summary`.
  - [x] `property-positions`.
  - [x] `class-lookup`.
  - [x] `class-property-summary`.
  - [x] `property-filter`.
  - [x] `structure-health`.
  - [x] `structure-owner-count`.
  - [x] `structure-owners`.
  - [x] `structure-owner-locations`.
  - [x] `structure-heatmap`.
  - [x] `base-components`.
  - [x] `dinos`.
  - [x] `dino-wild-tamables`.
  - [x] `dino-babies`.
  - [x] `dino-best-stat`.
  - [x] `dino-best-base-stat`.
  - [x] `dino-most-mutated`.
  - [x] `dino-wild-tamed`.
  - [x] `dino-heatmap`.
  - [x] `equipment-summary`.
  - [x] `equipment-saddles`.
  - [x] `equipment-best`.
  - [x] `equipment-rank`.
  - [x] `equipment-ascendant-weapon-bps`.
  - [x] `equipment-history`.
  - [x] `equipment-owned-by`.
  - [x] `stackables`.
  - [x] `stackable-owned-by`.
  - [x] `player-inventories`.
  - [x] `player-roster`.
  - [x] `tribe-roster`.
  - [x] `player-tribe-links`.
  - [x] `players`.
  - [x] `tribes`.
  - [x] `cluster`.
  - [x] `cluster-summary`.
  - [x] `tribute`.
  - [x] `export-json`.
  - [x] `export-domain-json`.
  - [x] `export-cluster-json`.
  - [x] `export-tribute-json`.
  - [x] Experimental `mutate`.
- [x] Test fixture cleanup:
  - [x] Centralize many public fixtures in `internal/testfixtures`.
  - [x] Continue migrating remaining domain-specific parser fixtures as touched;
        `arkapi` general/core synthetic save helpers now delegate header and
        object wrapping to `internal/testfixtures`, and `arkprofile` malformed
        archive tests now reuse shared archive framing/string/property writers.
        Dino/equipment string property payload writers and equipment positioned
        UInt16 property payload writers are now shared, and dino scalar,
        object-reference, positioned stat/color, and name-array property payload
        writers now use shared helpers; dino custom item data struct/byte-array
        payload writers now use shared helpers; low-level `arkproperty` and
        `arkobject` name-table-ID, integer, float, string, and compound
        custom-data test writers now use shared helpers; API malformed
        class-only object row fixtures and API/benchmark actor-transform custom
        table fixtures now use shared helpers; save-contained player/tribe game
        object fixture payloads now share the archive property builders while
        keeping game-object framing explicit; ID-table game-object bytes with
        custom object-name payloads now use a shared testfixtures helper;
        simple ID-table int-property object rows now use
        `testfixtures.ObjectBytesWithIntProperty` directly in
        save-layer and arkapi general/core tests without local wrapper functions;
        embedded `GameModeCustomBytes` player/tribe fixture assembly and
        minimal embedded cryopod archive test payloads now live in shared
        testfixtures, and API/object-model cryopod dino/saddle payload tests
        now call shared fixtures directly; save-layer malformed full-object
        truncation fixtures now use a shared helper; parsed
        `CustomItemDatas` cryopod/custom-data fixtures now live in
        `internal/propertyfixtures`, and binary `CustomItemDatas` writers now
        live in shared testfixtures; ID-table Vector struct property writers
        now live in shared testfixtures for example/player-location fixtures;
        base linked-structure object-reference array fixtures now use shared
        ID-table array writers; player/tribe relation directory fixtures are
        shared by API and CLI tests; dino stats/status object byte fixtures are
        shared by API and CLI tests; stackable API tests and benchmarks use
        shared stackable object fixtures directly; CLI archive and tribute
        smoke tests call shared archive/tribute file fixtures directly instead
        of local wrapper functions. Remaining `synthetic*`/`createSynthetic*`
        helpers are package-local malformed payloads, purpose-built name-table
        headers, or domain-specific save graphs that describe the behavior under
        test.
- [x] Regression: re-run `make verify` and focused private oracle comparisons
      after each committed behavior slice.

## Phase 4: Documentation And Production Readiness

Done when another engineer can build, test, run, and extend the offline parser
without Python or private chat context.

- [x] README covers install/build, CLI, library use, examples, scope, and
      mutation safety.
- [x] Supported file types documented: `.ark`, `.arkprofile`, `.arktribe`,
      local cluster files, and local tribute index files.
- [x] Unsupported features documented: FTP, RCON, live server integration, and
      unsupported legacy archive paths.
- [x] Mutation APIs documented as experimental and live-server-unverified.
- [x] Oracle regeneration documented for `~/Downloads/SavedArks.tar.bz2`.
- [x] Privacy rules and ignored paths documented.
- [x] Opt-in CLI redaction documented and tested.
- [x] Standalone Go examples added for implemented offline workflows.
- [x] Public Go packages have package-level `doc.go` documentation, including
      scope and safety notes for API, save, mutation, archive, profile,
      cluster, tribute, object, property, binary, and logging packages.
- [x] Verification commands documented.
- [x] `go test ./...` passes.
- [x] `go vet ./...` passes.
- [x] `make build` passes.
- [x] `make verify` passes.
- [x] Public GitHub Actions runs `make verify` on `main` pushes, pull
      requests, and manual dispatches without private oracle data. The README
      badge links to current status; historical passing runs include
      `28346712289` and `28347063999`.
- [x] CLI and example smoke tests pass on synthetic fixtures.
- [x] Static/local release binary builds with `CGO_ENABLED=0`.
- [x] Static/local release binary exposes build metadata through
      `arksave --version`.
- [x] Go-only provided-data E2E smoke tests cover selected read-only APIs, CLI
      commands, map/save metadata summaries, object class lists, object
      summaries, property-position metadata, class/property lookup and class
      property summaries, local
      profile/tribe/tribute file handling, and aggregate-output examples
      through `make e2e-test`; they skip without `ARK_E2E_SAVE` or
      `ARK_E2E_SAVE_DIR`. A 2026-06-29 provided-data rerun passed against the
      private provided Valguero save and directory under ignored `.oracle`.
- [x] Go-only provided-data E2E smoke tests cover stackable, stackable-owned,
      dino, and equipment command/API paths, including ascendant weapon
      blueprint, direct-dino base-stat and heatmap summaries, and
      equipment-owned-by summaries, plus domain JSON export through both
      `arkapi.JSONAPI.ExportDomain` and the CLI
      `export-domain-json` path.
- [x] CLI and example smoke tests cover generated equipment history manifests
      and explicit JSON report output on synthetic fixtures.
- [x] Go-only provided-data E2E smoke tests cover bounded structure and base
      read APIs through structure owner/health summaries, structure CLI
      commands, structure heatmap JSON export, structure owner/location
      examples, and selected-property base component stats without requiring
      full structure/base JSON export in the smoke gate.
- [x] Go-only provided-data E2E smoke tests cover typed local-cluster API and
      examples when `ARK_E2E_CLUSTER` or `ARK_E2E_CLUSTER_DIR` is configured.
- [blocked] Private oracle comparison suite exists and currently records
      forty-six passing sanitized comparison cases for selected implemented
      features. The Make target now gives upstream subprocesses access to the
      oracle venv `python`, and the 2026-06-29 rerun progressed past the
      previous malformed-cryopod/logger failure before being interrupted while
      CPU-bound in upstream structure parsing; no Go mismatch was produced
      before interruption. A focused `tribe_list` rerun passed through
      `make oracle-compare ORACLE_COMPARE_ARGS="--case tribe_list"` with the
      private provided save.
- [x] Expanding the private oracle comparison suite to every runnable upstream
      Python example is intentionally out of scope.
- [x] Final production-readiness review after Phase 4 docs, release build,
      smoke, provided-data, and oracle-comparison checks are refreshed.

## Monitor Commands

Use these commands to monitor progress from the repo:

```sh
make status
```

```sh
rg -n "^\\s*- \\[ \\]|\\[~\\]|\\[blocked\\]" docs/project-task-ledger.md docs/phase-*.md docs/production-readiness-review.md
```

```sh
make verify
```

```sh
ARK_ORACLE_SAVE="/path/to/private/save.ark" \
  make oracle-compare
```

The full private oracle comparison can be slow. Focused cases can be run with
`make oracle-compare ORACLE_COMPARE_ARGS="--case <case>"`, where the supported
focused cases are listed by the script argument parser.
