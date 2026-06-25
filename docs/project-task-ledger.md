# Project Task Ledger

This is the monitorable task list for the offline-only Go port of upstream
`ark-save-parser`. It exists so progress can be audited from the repository
without reading chat history.

Status markers:

- `[x]`: complete and committed.
- `[~]`: partially implemented or intentionally narrowed; see the note.
- `[ ]`: not complete.
- `[blocked]`: blocked by missing fixture data, upstream behavior, live-server
  validation, or an explicit out-of-scope boundary.

## Current Rollup

| Phase | Status | Done Means | Current Blocking Items |
| --- | --- | --- | --- |
| Phase 1: Oracle Setup | `[~]` Mostly complete | Python oracle behavior is reproducible from private data and every runnable offline test/example has a recorded status. | A few read-only/mutation example outputs are now classified or structurally covered, but not every upstream example has a private normalized output artifact. |
| Phase 2: Literal Go Transpilation | `[~]` In progress | Offline Go behavior mirrors runnable upstream behavior closely enough for oracle-derived tests/examples to pass or have documented blockers. | Remaining parser edge cases, legacy archive parsing, full domain/model parity, slow full-save examples, and blocked cryopod/pedigree oracle paths. |
| Phase 3: Idiomatic Go Refactor | `[~]` In progress | Translated behavior is organized into stable Go packages and CLI surfaces without losing oracle parity. | Domain model/package split and fixture migration are not complete. |
| Phase 4: Documentation And Production Readiness | `[~]` In progress | Another engineer can build, test, run, and extend the project without Python/private context. | Production readiness remains blocked by full runnable-example oracle coverage and full offline API/domain parity. |

## Operating Rules

- [x] Keep FTP and RCON out of scope.
- [x] Keep cluster support local-file-only.
- [x] Treat mutation APIs as experimental and live-server-unverified.
- [x] Keep private saves, oracle raw output, snapshots, debug dumps, private
      manifests, and extracted save data out of git.
- [x] Push coherent task-group commits to `main`.
- [x] Use sanitized docs for progress reporting; detailed private oracle values
      stay under `.oracle/output`.
- [ ] Update this ledger in the same commit whenever a new task, blocker, or
      completion status is discovered.

## Phase 1: Oracle Setup

Done when Python oracle behavior is reproducible from private backup data and
every runnable offline upstream test or example has a recorded sanitized status.

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
- [~] Run offline examples:
  - [x] Classify read-only, mutation-copy, export-producing,
        heatmap-producing, local-cluster, network-only skip, and impossible
        examples in `docs/upstream-oracle-classification.md`.
  - [x] Add privacy-safe oracle comparison harness in
        `scripts/oracle_compare.py`.
  - [x] Capture private oracle comparison output under `.oracle/output`.
  - [x] Commit sanitized aggregate status in
        `docs/oracle-comparison-summary.md`.
  - [~] Read-only examples are covered incrementally by Go counterparts and
        sanitized aggregate comparisons; not every upstream example has a
        default-suite comparison because some are slow or blocked.
  - [~] Mutation examples are structurally represented by copy/write/reopen
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
- [ ] Port remaining compound payload encodings discovered by future oracle
      failures.
- [ ] Port legacy property/object parsing where a runnable local oracle path
      exists.

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
- [ ] Complete remaining read-first wrappers for lower-priority inventory,
      owner, trait, dino, structure, equipment, stackable, player, tribe, and
      local cluster fields as oracle examples require them.

### Offline APIs

- [x] Port General API object queries and fault collection.
- [x] Port local profile and tribe discovery/loading.
- [x] Port save-contained player and tribe parsing, including game-table
      objects and embedded `GameModeCustomBytes`.
- [x] Port player lookups, tribe relationships, player aggregates,
      unlocked-engrams aggregation, pawn lookup, inventory lookup, and location
      lookup.
- [x] Port local tribute discovery/loading and JSON export.
- [x] Port local cluster discovery/loading, uploaded item/dino summaries, typed
      projections, parse-status reporting, and JSON export.
- [x] Port Structure API class, owner, location, connected-structure, heatmap,
      inventory-container, and selected-property fast scan paths.
- [x] Port Stackable API counts, categories, owner filtering, and JSON export.
- [x] Port Equipment API weapon/armor/saddle/shield queries, owner filtering,
      crafted state, quality/rating/durability filters, stat calculations,
      ranking helpers, and modern cryopod saddle extraction.
- [x] Port Dino API class/tamed/wild/baby/stat/mutation/trait/ownership/
      ancestry/pedigree-base/heatmap/filter helpers and modern cryopod
      dino/status extraction.
- [x] Port Base API grouping, point lookup, minimum count filtering, all-base
      discovery, and fast base-component statistics.
- [x] Port save-info and implemented domain JSON exports.
- [x] Add explicit-output `export_all_items` example and manifest.
- [x] Add local multi-save `equipment_history` example.
- [ ] Finish full dino edge behavior:
  - [ ] Legacy/modded cryopod variants.
  - [ ] Cryopod-location example parity when upstream/private data permits.
  - [ ] Full pedigree rendering/tree export parity beyond JSON child and
        descendant references.
- [ ] Finish full structure/base edge behavior:
  - [ ] Exact `structure_owner_locations` owner/cell parity without
        prohibitive full-parse runtime.
  - [blocked] `structure_heatmap` oracle comparison, blocked because upstream
        indexes out-of-range cells on the supplied private save.
  - [ ] Base export/import read/write parity where local-copy structural tests
        are feasible.
- [ ] Finish full equipment edge behavior:
  - [ ] Exact equipment ranking count and average-stat parity.
  - [ ] Legacy/modded cryopod saddle payloads and cosmetics.
  - [ ] Remaining default armor/stat table parity.
- [ ] Finish richer local cluster item/dino domain models as new local-file
      oracle fixtures expose fields.
- [ ] Finish remaining Player/Tribe edge behavior not covered by parsed local
      archives, game-table objects, or embedded `GameModeCustomBytes`.
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
- [~] Upstream generated-blueprint insertion is classified as mutation-copy
      only; live-server acceptance remains unverified.
- [~] Upstream base import/customize examples have partial structural coverage
      through exported raw structure rows and `ImportBaseBinary` reinsert into
      explicit copied saves. Moving structures, inventory expansion,
      customization, owner replacement, and live-server acceptance remain
      unverified.
- [~] Upstream dino extract/reinsert examples have partial structural coverage
      through exported direct-save dino, status, and inventory rows plus
      `ImportDinoBinary` reinsert into explicit copied saves. Cryopod insertion
      into target inventories, generated location changes, stat mutation, and
      live-server acceptance remain unverified.
- [~] Upstream generated-blueprint/equipment insertion example has partial
      structural coverage through exported equipment rows and
      `ImportEquipmentBinary` reinsert into explicit copied saves. Generated
      blueprint construction, insertion into target inventories, and
      live-server acceptance remain unverified.
- [ ] Add more structural mutation tests for upstream dino trait/stat/growth
      and structure mutation examples where local-copy behavior can be
      represented without claiming live-server safety.

### Examples And Oracle Comparisons

- [x] `map_summary`.
- [~] `parse_all`: implemented and smoke-tested; private comparison is manual
      because full save parsing is runtime-heavy.
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
- [blocked] `dino_cryopod_location`: upstream/private malformed cryopod path
      blocks stable oracle output.
- [blocked] `dino_pedigrees`: upstream/private malformed cryopod path blocks
      stable oracle output.
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
- [x] `base_components`.
- [x] `base_export_from_save`: represented as an explicit-output structural
      export that writes base metadata, copied raw structure rows, and
      structure location JSON. Inventory item expansion, binary import, and
      live-server validation remain mutation-copy-adjacent and unverified.
- [x] `cluster_json`.
- [x] `local_tribute`.
- [x] `tribute_json`.
- [x] `export_json`.
- [~] `export_all_items`: implemented and smoke-tested; default oracle
      comparison is deferred because full export is too slow on the large
      private save.
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
- [~] API cleanup:
  - [x] Add typed APIs for implemented player, tribe, dino, structure,
        equipment, stackable, base, cluster, tribute, and JSON workflows.
  - [x] Use explicit errors and fault collections for partial object parses.
  - [x] Require explicit output paths for mutation operations.
  - [ ] Split or further organize large domain models under `arkobject` or
        subpackages once behavior stabilizes.
  - [ ] Replace remaining Python-shaped compatibility helpers where typed Go
        surfaces now exist.
- [~] Performance pass:
  - [x] Add benchmarks for full save open/object enumeration, object parse,
        query filters, and JSON export.
  - [x] Add selected-property scans for expensive structure/base workflows.
  - [ ] Add object cache controls where benchmarks show repeat parsing is a
        practical bottleneck.
  - [ ] Add safe concurrency only where tests prove no behavior drift.
- [x] CLI:
  - [x] `inspect`.
  - [x] `parse`.
  - [x] `players`.
  - [x] `tribes`.
  - [x] `cluster`.
  - [x] `tribute`.
  - [x] `export-json`.
  - [x] `export-domain-json`.
  - [x] `export-cluster-json`.
  - [x] `export-tribute-json`.
  - [x] Experimental `mutate`.
- [~] Test fixture cleanup:
  - [x] Centralize many public fixtures in `internal/testfixtures`.
  - [ ] Continue migrating remaining domain-specific parser fixtures as touched.
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
- [x] Verification commands documented.
- [x] `go test ./...` passes.
- [x] `make build` passes.
- [x] `make verify` passes.
- [x] CLI and example smoke tests pass on synthetic fixtures.
- [x] Static/local release binary builds with `CGO_ENABLED=0`.
- [~] Private oracle comparison suite exists and currently records forty-six
      passing sanitized comparison cases.
- [ ] Private oracle comparison suite covers every runnable upstream Python
      example that is feasible on available data and reasonable runtime.
- [ ] Final production-readiness review after Phase 2 and Phase 3 remaining
      gaps are closed.

## Monitor Commands

Use these commands to monitor progress from the repo:

```sh
rg -n "^- \\[ \\]|\\[~\\]|\\[blocked\\]" docs/project-task-ledger.md docs/phase-*.md docs/production-readiness-review.md
```

```sh
make verify
```

```sh
ARK_ORACLE_SAVE="$PWD/.oracle/data/SavedArks/Valguero_WP/Valguero_WP.ark" \
  .oracle/venv/bin/python scripts/oracle_compare.py
```

The full private oracle comparison can be slow. Focused cases can be run with
`scripts/oracle_compare.py --case <case>`, where the supported focused cases
are listed by the script argument parser.
