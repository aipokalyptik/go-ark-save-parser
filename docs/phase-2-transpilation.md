# Phase 2 Literal Go Transpilation

Phase 2 is complete when the Go code mirrors upstream offline behavior closely
enough that oracle-derived tests can run against translated packages.

## Requirements

- Preserve upstream behavior first, even when the shape is not idiomatic Go.
- Keep FTP and RCON out of scope.
- Support local `.ark`, `.arkprofile`, `.arktribe`, local cluster files, and
  local tribute index files where fixtures exist.
- Treat mutation APIs as experimental and structurally tested only.
- Add tests before implementation code for each behavior slice.
- Keep private oracle data under `.oracle/`; committed tests must use synthetic
  fixtures or sanitized expected values.

## Task Ledger

### Core Binary Layer

- [x] Add tests for byte position, bounds handling, seek/skip, peek, and
      remaining-byte behavior.
- [x] Add tests for little-endian signed/unsigned integers, floats, doubles, and
      bools.
- [x] Add tests for ARK strings: positive ASCII/null-terminated, empty strings,
      and negative UTF-16LE strings.
- [x] Add tests for UUID byte order.
- [x] Add tests for name-table lookup and missing-name behavior.
- [x] Add tests for zlib inflation wrapper behavior.
- [x] Add tests for ARK wildcard decompression edge cases.
- [x] Implement the minimal binary reader/writer to pass those tests.

### Save Access

- [x] Add synthetic SQLite `.ark` fixture tests with `custom` and `game` tables.
- [x] Port save context, header parsing, name table parsing, custom table reads,
      and object binary access.
- [x] Parse `ActorTransforms` custom table into save context.
- [x] Add private-oracle integration tests gated behind an environment variable.
- [x] Validate object enumeration against oracle data.

### Property Parser

- [x] Add tests for property terminator handling.
- [x] Add tests for primitive property types.
- [x] Add tests and parsing for raw byte and enum `ByteProperty` encodings.
- [x] Add tests for unknown struct
      fallback behavior.
- [x] Add tests and parsing for object-reference properties and simple value arrays.
- [x] Add tests and parsing for ASA local-cluster primitive edge types:
      `SoftObjectProperty`, `NameProperty`, `Int8Property`, `Int16Property`,
      `Int64Property`, `UInt16Property`, and inline `UInt64Property` values.
- [x] Add tests and generic parsing for struct arrays.
- [x] Add tests and generic property-list parsing for struct properties.
- [x] Add tests and parsing for simple value maps and sets.
- [x] Add tests and parsing for map entries with generic struct values.
- [x] Add tests and parsing for nested map/set property-list edge cases seen in
      structure wireless exchange references.
- [x] Add raw fallback preservation for packed unknown structs.
- [x] Add packed `Vector` struct parsing for pawn/world-location properties.
- [x] Add declared-size realignment for parsed primitive property payloads.
- [x] Add declared-size realignment and overread detection for compound array,
      map, set, and struct payloads covered by generic readers.
- [ ] Port remaining property parsing edge cases for compound payload encodings
      not already covered by generic readers.
- [x] Isolate legacy archive behavior behind explicit format paths and a typed
      unsupported-legacy error.
- [ ] Port legacy property/object parsing where a runnable offline oracle path exists.

### Object Model

- [x] Port generic game object headers and property containers.
- [x] Add read-first inventory and inventory item wrappers for parsed objects.
- [x] Add read-first object owner wrapper for parsed ownership fields.
- [x] Port actor transforms and map coordinate helpers.
- [x] Add read-first structure wrapper for parsed object fields.
- [x] Add read-first equipment wrapper for parsed inventory item fields.
- [x] Add read-first object crafter metadata wrapper for crafted inventory/equipment
      items.
- [x] Add read-first player and tribe wrappers for parsed profile containers.
- [x] Add read-first dino wrapper for parsed object identity and status fields.
- [x] Add read-first base summary wrapper for grouped structures.
- [x] Add read-first structure inventory metadata for `MyInventoryComponent`,
      `CurrentItemCount`, `MaxItemCount`, open slot, and empty-container queries.
- [ ] Port remaining inventory, owner, trait, dino, structure, equipment,
      stackable, player, tribe, and local cluster data models as read-first wrappers.
- [x] Preserve raw binary/property positions and encoded byte spans needed by
      mutation structural tests.

### Offline APIs

- [x] Port General API object queries.
- [x] Add broad General API object parsing with fault collection so callers can
      keep valid parsed objects while inspecting per-object parse failures.
- [x] Add archive metadata parser and local `.arkprofile` / `.arktribe` file-open wrappers.
- [x] Parse modern local archive object properties on a best-effort basis with a
      strict mode for debugging unsupported property encodings.
- [x] Add first normalized tribe summary extraction surface.
- [x] Add offline CLI metadata summaries for local `.arkprofile` and `.arktribe` files.
- [x] Add first local-file Player/Tribe API surface for profile and tribe discovery/loading.
- [x] Add read-only local profile parsing into normalized player summaries.
- [x] Add read-only local tribe summary parsing through Player API and CLI output.
- [x] Add typed local player/tribe lookup helpers for player data IDs, tribe IDs,
      and parsed tribe detail objects.
- [x] Add typed local Player/Tribe API filters for player names, unique IDs,
      tribe names, owner IDs, and tribe member names/IDs.
- [x] Add typed local Player/Tribe relationship helpers for tribe-to-player
      maps, tribe-of-player lookup, and profile-derived object/dino owners.
- [x] Add typed local Player API death-stat helpers for per-player death
      counts, total deaths, and highest-death player lookup.
- [x] Add typed local Player API level, experience, and engram point parsing
      plus level, experience, and engram point lookup helpers.
- [x] Add typed local Player API aggregate helpers for total experience plus
      average level and experience.
- [x] Add typed local Player API lowest-stat lookups for deaths, level, and
      experience.
- [x] Add typed local Player API aggregate helpers for average deaths, total
      level, and lowest engram-point lookup.
- [x] Add read-only local Player API unlocked-engram blueprint aggregation for
      upstream unlocked-engrams example parity.
- [x] Add save-contained Player API pawn lookup by linked player data ID as the
      first read-only player inventory/location prerequisite.
- [x] Add save-contained Player API inventory lookup by linked player data ID,
      resolving pawn inventory components to inventory item UUID lists.
- [x] Add save-contained Player API location lookup by linked player data ID,
      resolving pawn `SavedBaseWorldLocation` vectors.
- [x] Add typed local Tribe API dino-count aggregate helpers.
- [x] Include local tribute index discovery/loading in the directory-based
      Player API surface.
- [x] Add first local-file cluster archive discovery/loading surface for extensionless
      cluster files.
- [x] Add read-only local cluster item/dino payload extraction, including
      `ArkItems`, `ArkTamedDinosData`, upload metadata, item blueprint summaries,
      and best-effort cluster dino archive parsing.
- [x] Add richer local cluster uploaded item metadata for quantity, rating,
      quality index, and crafter names where present.
- [x] Add offline CLI summary for local cluster files and directories.
- [x] Port local-file Player and Tribe APIs for parsed profiles/tribes,
      directory discovery, local cluster/tribute indexing, lookup filters, and
      aggregate helpers.
- [x] Add CLI directory summaries for local player profiles and tribe saves.
- [x] Port save-contained Player and Tribe API parsing for `.ark` game-table
      `PrimalPlayerDataBP` and `PrimalTribeData` objects, including lookup,
      tribe-player relation, and owner helper reuse.
- [ ] Port any remaining upstream Player/Tribe edge behavior not covered by
      parsed `.arkprofile`, `.arktribe`, or save-contained player/tribe objects.
- [x] Add first read-only Structure API surface for class, owner, and location
      queries with optional class filters.
- [x] Add read-only Stackable API surface for local resource/ammo/consumable counts.
- [x] Add read-only Stackable API category filters for resources, ammo, and
      consumables.
- [x] Add first fault-tolerant domain API paths with `StackableAPI.AllWithFaults`,
      `StructureAPI.AllWithFaults`, `EquipmentAPI.AllWithFaults`, and
      `DinoAPI.AllWithFaults`, plus structure-derived `BaseAPI.AllWithFaults`,
      preserving valid parsed objects while reporting matching object parse faults.
- [x] Add save-contained `PlayerAPI.PlayersWithFaults` and
      `PlayerAPI.TribeDetailsWithFaults` for partial-success player and tribe
      scans over `.ark` player-data and tribe-data objects.
- [x] Narrow save-contained player pawn lookup to pawn classes so inventory and
      location helpers are not blocked by unrelated unsupported objects.
- [x] Parse `UniqueNetIdRepl` structs and recover partial nested profile structs
      so local `.arkprofile` player deaths, character names, stats, and unlocked
      engrams match the Python oracle aggregate behavior.
- [x] Add read-only save-info JSON export API and CLI command.
- [x] Add read-only domain JSON export API and CLI command for implemented
      dino, structure, equipment, stackable, and base summaries.
- [x] Include equipment crafted state, implemented stat names, and average
      internal stat ranking values in equipment domain JSON export.
- [x] Include structure inventory UUID, item counts, open slots, and empty-state
      metadata in structure domain JSON export.
- [x] Include stackable owner inventory UUIDs in stackable domain JSON export
      for owner/container correlation.
- [x] Include equipment owner inventory UUIDs in equipment domain JSON export
      for owner/container correlation.
- [x] Add upstream canonical equipment class-list coverage for blueprints that
      do not match broad weapon, saddle, armor, or shield path heuristics.
- [x] Add pre-parse Equipment API blueprint predicate composition for callers
      that need upstream-style filtered equipment scans.
- [x] Include base keystone and averaged map coordinates in base domain JSON
      export for upstream base-location example parity.
- [x] Route structure domain JSON export through the fault-tolerant structure
      scan so malformed matching rows do not abort valid exports.
- [x] Include parsed tamed dino owner/detail and baby maturation fields in
      dino domain JSON export.
- [x] Add read-first dino color-set indices/names and uploaded-server origin
      fields using upstream positioned-property semantics.
- [x] Add read-first linked dino status-component stats for base/current level,
      base/tamed/mutated stat points, current stat values, imprinting percent,
      and dino JSON export.
- [x] Add typed read-first dino gene trait parsing while preserving raw trait
      strings for compatibility.
- [x] Add typed Dino API filters for current level and combined/base/mutated
      stat point thresholds.
- [x] Add typed Dino API gene trait filters by trait name and optional level.
- [x] Add read-only Dino API aggregate helpers for level, class, and tamed
      state counts over filtered dino maps.
- [x] Add upstream-compatible object/dino short-name extraction and dino
      short-name count helpers without changing full-blueprint counts.
- [x] Add upstream-style cryopodded dino aggregate counts by class with an
      overall `all` total.
- [x] Add combined read-only Dino API filtering for level bounds, class names,
      tamed state, gene traits, cryopodded state, and stat minimums.
- [x] Add upstream-compatible tamed dino generation and ancestor ID extraction
      from female/male ancestor arrays; full pedigree graph reconstruction
      remains pending.
- [x] Add read-only childless tamed dino filtering from parsed ancestor IDs;
      full pedigree graph reconstruction remains pending.
- [x] Add upstream-style baby dino filtering flags for tamed, wild, and
      cryopodded inclusion.
- [x] Add read-only Dino API lookup by two-part dino ID with explicit wild
      inclusion.
- [x] Add read-only tamed dino ownership filtering by tribe target team.
- [x] Add read-only wild-tamed dino helper/filter parity for tamed dinos with
      no parsed ancestor records.
- [x] Add read-only wild-by-class and tamed-by-class Dino API helpers.
- [x] Add read-only wild-tamable Dino API helpers using upstream non-tameable
      blueprint classes.
- [x] Add read-only wild/tamed minimum-level Dino API helpers.
- [x] Add read-only tamed dino container lookup by inventory UUID.
- [x] Add read-only local cluster JSON export API and CLI command for single
      cluster files and directories.
- [x] Add first read-only Equipment API surface for weapon/armor/saddle/shield queries.
- [x] Add read-only Equipment API convenience wrappers for weapons, armor,
      saddles, shields, and quantity counting.
- [x] Add read-only Equipment API filtering by crafted item crafter metadata.
- [x] Add read-only Equipment API filtering by equipped/blueprint state,
      quality index, minimum rating, and minimum durability.
- [x] Add combined read-only Equipment API filtering for kind, class,
      blueprint state, quality, rating, durability, equipped state, and crafter.
- [x] Add read-only Equipment API kind-scoped class filtering for upstream
      `get_by_class(cls, classes)` parity.
- [x] Add read-first equipment `ItemStatValues` parsing for internal stat values
      plus upstream-compatible weapon damage and durability JSON export.
- [x] Add read-first armor `ItemStatValues` calculations for armor,
      hypothermal resistance, hyperthermal resistance, and JSON export.
- [x] Add typed Equipment API filters for parsed damage, armor, hypothermal
      resistance, and hyperthermal resistance thresholds.
- [x] Add typed Equipment API filtering by actual durability calculated from
      `ItemStatValues`, distinct from saved current durability percentage.
- [x] Add equipment model helpers for crafted detection, implemented stat lists,
      and upstream-compatible internal average-stat ranking.
- [x] Add first read-only Dino API surface for local class, tamed/wild,
      baby/adult, sex, and alive/dead queries.
- [x] Add read-first tamed dino details for tamed name, neuter state,
      inventory UUID, and owner/tamer/imprinter fields.
- [x] Add read-first baby dino maturation percentage and upstream-compatible
      baby/juvenile/adolescent stage classification.
- [x] Add typed dino stat ranking and mutation-count helpers for read-only
      equivalents of upstream best-stat and most-mutated-tamed example flows.
- [x] Add structure container lookup by inventory UUID and stackable owner
      filtering/counting through owning structure inventories.
- [x] Add read-only Structure API lookup by object UUID for upstream
      `get_by_id` parity.
- [x] Add read-only Structure API connected-structure traversal from parsed
      `LinkedStructures` UUID references.
- [x] Add read-only Structure API subset location filtering for upstream
      `filter_by_location` parity.
- [x] Add read-only Structure API subset owner filtering for upstream
      `filter_by_owner` parity, including its invert behavior.
- [x] Add read-only Structure API owned-structure counting and class-plus-owner
      filtering for upstream owned-structure and owned-vault example flows.
- [x] Add read-only Structure API heatmap generation with class, owner, and
      minimum-cell filters for upstream `create_heatmap` parity.
- [x] Add read-only Structure API discovery of missed inventory-bearing
      container structures plus engram filtering for upstream `get_all_objects`
      parity.
- [x] Add equipment owner filtering/counting through owning structure
      inventories for upstream owner-of-items example flows.
- [x] Add first read-only Base API surface for nearby owned structure grouping.
- [x] Expand read-only Base API point lookups through linked structures for
      upstream `get_base_at` parity.
- [x] Add read-only Base API filtering by minimum grouped structure count.
- [x] Add optionized read-only Base API all-base discovery for upstream
      `get_all_bases` parity, including connected-only and default minimum
      structure filtering.
- [ ] Port remaining full Dino edge behavior, Structure, Equipment, Stackable,
      Base, richer local cluster item/dino domain models, and remaining
      model-specific JSON API edge behavior.
- [x] Port compact `.arktributetribe` / `.arktributetribetribe` local tribute
      index parsing for player-data and tribe-data ID lists.
- [ ] Port legacy archive object parsing for any other runnable local oracle path
      that is not covered by modern archive or compact tribute index formats.
- [x] Mark unsupported FTP/RCON examples as skipped in compatibility docs.

### Experimental Mutation

- [x] Port first copy-based modification helpers where upstream behavior can be
      translated safely: copy save, remove object row, upsert object bytes, and
      upsert custom values on copied SQLite saves.
- [x] Require explicit output paths.
- [x] Add structural write/reopen/reparse tests only.
- [x] Document live-server validation as out of scope.

### Examples And Review

- [x] Add first Go examples for runnable offline Python example categories:
      map summary, class listing, local profile/tribe discovery, cluster JSON,
      and mutation-copy workflows.
- [x] Compare normalized Go outputs with the private Python oracle for the
      implemented direct read-only counterparts: `map_summary` and
      `object_classes`.
- [x] Compare normalized Go `local_profiles` aggregate counts with the private
      Python oracle using upstream `PlayerApi` over local profile/tribe files.
- [x] Compare normalized Go CLI `export-json` save-info output with the private
      Python oracle for save metrics and object class lists.
- [x] Compare normalized Go `property_filter` aggregate counts with the
      upstream property-name prefilter workflow.
- [x] Compare normalized Go `stackable_count` aggregate counts with the
      upstream `StackableApi.get_by_class` and `get_count` workflow.
- [x] Compare normalized Go equipment domain JSON aggregates with the upstream
      longneck blueprint max-damage workflow.
- [x] Compare normalized Go `cluster_json` aggregate counts with the upstream
      Python `ClusterData` parser over upstream local cluster fixture files.
- [x] Compare normalized Go `local_tribute` aggregate counts with private local
      compact tribute index files.
- [ ] Compare normalized Go outputs with private Python oracle outputs where
      available for the remaining runnable upstream examples.
- [x] Run subagent spec and quality reviews on parser parity and API coverage;
      current findings are recorded in `docs/production-readiness-review.md`.
