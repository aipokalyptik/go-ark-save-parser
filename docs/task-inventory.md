# Task Inventory

This is the up-front, stable task inventory for the offline-only Go port. It is
intended for progress monitoring without reading chat history. Detailed status
notes live in [`project-task-ledger.md`](project-task-ledger.md); phase-specific
evidence lives in the linked phase documents.

Status markers:

- `[x]`: complete and committed.
- `[~]`: partially complete; the remaining work is listed in the same row.
- `[ ]`: not complete.
- `[blocked]`: blocked by fixture availability, upstream behavior, runtime
  limits, live-server validation, or explicit scope boundaries.

## Scope Rules

| ID | Status | Requirement | Evidence / Remaining Work |
| --- | --- | --- | --- |
| SCOPE-001 | `[x]` | Offline local-file access is the compatibility target. | README scope and package docs cover `.ark`, `.arkprofile`, `.arktribe`, local cluster, and local tribute files. |
| SCOPE-002 | `[x]` | FTP is unsupported. | Documented as intentionally out of scope. |
| SCOPE-003 | `[x]` | RCON is unsupported. | Documented as intentionally out of scope. |
| SCOPE-004 | `[x]` | Cluster support is local-file-only. | `arkcluster` and CLI cluster commands read local extensionless archive files only. |
| SCOPE-005 | `[~]` | Mutation APIs are translated only as explicit copied-save workflows. | `arkmutation` requires output paths and has structural tests; live-server acceptance remains unverified by design. |
| SCOPE-006 | `[x]` | Private saves, raw oracle output, debug dumps, extracted saves, and snapshots never enter git. | `.gitignore`, development docs, and sanitized oracle docs define the privacy boundary. |
| SCOPE-007 | `[x]` | Do not expand the Python codebase or oracle suite beyond selected-feature parity needs. | Existing oracle artifacts remain reference evidence only; new work should target Go offline feature parity and chosen examples. |
| SCOPE-008 | `[x]` | Expand Go code coverage, examples, and provided-data E2E tests for chosen offline features. | `make e2e-test` runs Go-only read-only API, stackable/dino/equipment domain JSON API/CLI export, bounded structure/base API checks, CLI profile/tribe/tribute file handling, and examples smoke paths against `ARK_E2E_SAVE` or `ARK_E2E_SAVE_DIR`. |

## Phase 1: Oracle Setup

| ID | Status | Requirement | Evidence / Remaining Work |
| --- | --- | --- | --- |
| P1-001 | `[x]` | Initialize repo hygiene. | `go.mod`, `.gitignore`, MIT license, public GitHub repository, and pushed `main` exist. |
| P1-002 | `[x]` | Prepare Python oracle workspace from private backup. | `.oracle` workflow documented in [`development.md`](development.md) and Phase 1 docs. |
| P1-003 | `[x]` | Clone upstream at commit `4f7cc91fb96a080321bfbc884ba81bd897f72c49`. | Oracle setup docs pin the upstream commit. |
| P1-004 | `[x]` | Use Python 3.13+ and install upstream editable package plus test tooling. | Oracle regeneration docs record the setup commands. |
| P1-005 | `[x]` | Inventory `.ark`, `.arkprofile`, `.arktribe`, local cluster, and local tribute files privately. | Sanitized count-only summary exists in [`oracle-summary.md`](oracle-summary.md). |
| P1-006 | `[x]` | Run upstream packaged tests and record blockers. | Missing non-public upstream `tests/test_data` is documented. |
| P1-007 | `[x]` | Run upstream `testbench/pytest` against usable private `.ark` saves. | Sanitized status is recorded in Phase 1/oracle docs. |
| P1-008 | `[x]` | Classify offline upstream examples for selected-feature parity. | Classification and comparison harness exist; expanding Python/oracle coverage for every upstream example is no longer a project goal. |
| P1-009 | `[x]` | Capture Python-vs-Go oracle output privately for selected implemented features. | Forty-six aggregate comparison cases are recorded; existing oracle outputs are reference evidence, not an expansion target. |
| P1-010 | `[x]` | Review oracle completeness and privacy boundaries. | Phase 1 report and privacy docs are committed. |

## Phase 2: Literal Go Transpilation

| ID | Status | Requirement | Evidence / Remaining Work |
| --- | --- | --- | --- |
| P2-BIN-001 | `[x]` | Port byte reader/writer behavior. | `arkbinary` tests cover bounds, seek, skip, peek, and remaining bytes. |
| P2-BIN-002 | `[x]` | Port strings, UUIDs, numeric values, bools, arrays, structs, name tables, and position semantics where exercised. | `arkbinary`, `arkproperty`, and `arkobject` tests cover implemented encodings. |
| P2-BIN-003 | `[x]` | Port zlib inflation and ARK wildcard decompression. | `arkbinary` tests cover both paths. |
| P2-SAVE-001 | `[x]` | Port SQLite `.ark` loading using pure-Go SQLite. | `arksave` uses `modernc.org/sqlite` and synthetic SQLite tests. |
| P2-SAVE-002 | `[x]` | Port save header, custom tables, actor transforms, class lookup, object binary access, and object enumeration. | `arksave` tests and gated private oracle tests. |
| P2-PROP-001 | `[x]` | Port primitive, object, soft object, name, byte, enum, array, map, set, and generic struct property parsing. | `arkproperty` tests cover the implemented parser surface. |
| P2-PROP-002 | `[x]` | Preserve unknown property/struct fallback and declared-size realignment. | `arkproperty` tests cover fallback, recoverable overread behavior, and partial-parser continuation after aligned malformed compound properties. |
| P2-PROP-003 | `[ ]` | Port future compound payload encodings discovered while implementing chosen offline features. | No open chosen-feature failure currently has a complete parser slice; add cases when Go feature work exposes them. |
| P2-PROP-004 | `[ ]` | Port legacy property/object parsing where a runnable offline oracle path exists. | Legacy paths are isolated behind typed unsupported errors until fixtures are available. |
| P2-OBJ-001 | `[x]` | Port generic game objects and raw position/span preservation. | `arkobject` and mutation tests cover parsing and structural spans. |
| P2-OBJ-002 | `[x]` | Port read-first wrappers for inventory, owner, structure, equipment, player, tribe, dino, base, and cluster summaries. | Implemented wrappers live under `arkobject` and `arkcluster`. |
| P2-OBJ-003 | `[~]` | Complete remaining read-first wrappers as oracle examples require them. | Player and tribe models now expose typed name/membership helpers used by API summaries; lower-priority fields for inventory, traits, dino, structure, equipment, stackable, player, tribe, and local cluster remain incremental. |
| P2-API-001 | `[x]` | Port General API queries and fault collection. | `arkapi.GeneralAPI` tests. |
| P2-API-002 | `[x]` | Port local profile, tribe, local tribute, and local cluster parsing. | `arkprofile`, `arktribute`, `arkcluster`, and `arkapi` tests. |
| P2-API-003 | `[x]` | Port save-contained player and tribe parsing, including embedded `GameModeCustomBytes`. | `arkapi` player/tribe tests. |
| P2-API-004 | `[x]` | Port Dino, Structure, Equipment, Stackable, Base, and JSON read APIs for implemented offline workflows. | Domain tests, examples, and CLI/domain JSON tests, including provided-data E2E for stackable, dino, and equipment JSON exports plus bounded structure/base API checks. |
| P2-API-005 | `[~]` | Finish full dino edge behavior. | Typed pedigree tree helpers, domain JSON pedigree trees, typed malformed cryopod dino payload faults, and typed unsupported cryopod saddle payload faults exist; legacy/modded cryopod parsing and cryopod-location parity remain. Upstream/private full pedigree oracle comparison remains blocked by malformed cryopod path. |
| P2-API-006 | `[~]` | Finish full structure/base edge behavior. | Generated base/structure binary exports round-trip through mutation imports; typed structure health summaries and owner-location skip counters are covered by Go fixtures; exact owner/cell parity, base customization write parity, and live-server acceptance remain. Structure heatmap oracle is blocked by upstream private-save cell indexing. |
| P2-API-007 | `[~]` | Finish full equipment edge behavior. | Ranking candidate selection, high-rating aggregate stats, inventory-state summaries, upstream-list classification guards, upstream family/slot default stat tables, cursed shield defaults, inferable cursed weapon defaults, Tek sword durability, generic `CustomItemDatas` metadata, and typed unsupported cryopod saddle fault handling now have tested Go helpers; remaining long-tail default stat-table parity, legacy/modded cryopod saddle parsing, and exact cosmetic semantics remain. |
| P2-API-008 | `[~]` | Finish richer local cluster item/dino domain models. | Uploaded item type has typed constants/accessors, enum filters, version/parse helpers, crafter helpers, short names, item aggregate summaries, explicit uploaded-dino parse status counts, dino parse/component aggregate summaries, primary/short dino class names, JSON parse/version/component status fields, and directory-level aggregate summaries exposed through JSON and CLI output; cluster CLI rows also show item short names, dino primary/short class names, and dino parse-status aggregate counts. Add deeper item/dino fields only when chosen local-file fixtures expose them. |
| P2-API-009 | `[~]` | Finish remaining Player/Tribe edge behavior. | Typed player pawn inventory indexing, upstream-style inventory item counting, save-backed inventory summaries, path-level player inventory summaries, model-backed roster/data summaries, player-all aggregate summaries, player/tribe relationship edge counters, and local player death/level/experience average helpers now live in `arkapi` and examples; remaining upstream edge cases beyond parsed local archives, save objects, and embedded `GameModeCustomBytes` remain. |
| P2-MUT-001 | `[x]` | Port copy-based DB modification, object removal, object upsert, and custom-table upsert. | `arkmutation` tests and CLI mutate commands. |
| P2-MUT-002 | `[~]` | Translate higher-level mutation examples where feasible. | Generated base/structure/dino/equipment binary exports round-trip through mutation imports into reopenable copied saves; generated blueprint/base customization live-server acceptance is unverified. |
| P2-EX-001 | `[x]` | Create Go equivalents for runnable offline Python examples that currently have implemented API support. | `examples/` contains committed Go examples and smoke tests. |
| P2-EX-002 | `[x]` | Compare Go example output to Python oracle output for selected implemented examples. | Forty-six aggregate cases pass; expanding Python/oracle coverage for every feasible upstream example is no longer a project goal. |

## Phase 3: Idiomatic Go Refactor

| ID | Status | Requirement | Evidence / Remaining Work |
| --- | --- | --- | --- |
| P3-PKG-001 | `[x]` | Split packages for binary, save, property, object, profile, cluster, tribute, API, mutation, and CLI surfaces. | Current package shape is documented in [`phase-3-refactor.md`](phase-3-refactor.md). |
| P3-PKG-002 | `[~]` | Further split large domain models under `arkobject` or subpackages. | Shared crafter metadata now lives in its own `arkobject` file; broader dino, structure, equipment, stackable, player, tribe, inventory, and cluster subpackage splits remain. |
| P3-API-001 | `[x]` | Replace dynamic returns with typed Go structs, explicit errors, and fault collections for implemented paths. | Implemented typed APIs and domain JSON exports. |
| P3-API-002 | `[~]` | Replace remaining Python-shaped compatibility helpers where typed Go surfaces now exist. | `parse_all`, `object_classes`, `object_summary`, `class_property_summary`, `class_lookup`, `property_filter`, and `property_positions` now use typed `arkapi.GeneralAPI` helpers; player/tribe examples now use `arkapi.NewPlayerFromPath`; `player_all`, `player_list`, `tribe_list`, `player_tribe_links`, and `player_and_tribe_data` now use typed roster/relation/aggregate summaries; `local_profiles` now uses typed local-file and tribe-player link summaries; `local_tribute` now uses typed tribute file/directory summaries; `player_inventories` now uses `arkapi.PlayerInventorySummaryFromPath`; equipment max-damage, equipment best-item helpers, equipment history snapshots/report assembly, ascendant weapon blueprint summaries, equipment owned-by, saddle max-armor, equipment saddle totals with cryopod saddles, equipment summary with cryopod saddles, canonical equipment blueprint-list composition, wild-tamed max-level, dino stat token helpers, dino and structure heatmap summaries, dino population summaries, dino baby wild/tamed summaries, wild tamable counts, stackable count/owned, structure owner count, structure location examples, and all-domain JSON exports now use typed `arkapi` helpers; remaining generic inspection and mutation examples keep low-level access where that is the feature. |
| P3-API-003 | `[ ]` | Add remaining full typed API layers and model-specific JSON exports. | Stackable count/owner-filter, structure tribe-ownership, structure health, structure heatmap summary, dino wild-tamable, equipment inventory-state, equipment owned-by summary, equipment saddle summary including cryopod saddles, equipment summary including cryopod saddles, player-all aggregate, and shared heatmap summaries now have typed API helpers; full dino, full structure, equipment, base, cluster, and player/tribe parity remain incremental. |
| P3-PERF-001 | `[x]` | Add benchmarks for full save open/object enumeration, object parse, query filters, and JSON export. | Benchmarks are committed. |
| P3-PERF-002 | `[x]` | Add object cache controls and prove safe concurrency only where tested. | `arksave.Save` object row cache and concurrent raw read tests exist. |
| P3-CLI-001 | `[x]` | Implement offline CLI commands. | `inspect`, `parse`, `structure-health`, `structure-owner-count`, `structure-owners`, `structure-owner-locations`, `base-components`, `dinos`, `dino-wild-tamables`, `dino-babies`, `equipment-summary`, `equipment-saddles`, `player-inventories`, `player-roster`, `tribe-roster`, `player-tribe-links`, `players`, `tribes`, `cluster`, `cluster-summary`, `tribute`, JSON export commands, and experimental mutate commands exist. |
| P3-FIX-001 | `[~]` | Replace duplicated synthetic fixture builders with internal helpers. | `internal/testfixtures` centralizes many shared fixtures, including player pawn, inventory, structure, equipment, stackable, basic dino save-object builders, binary `CustomItemDatas` writers, ID-table Vector struct property writers, and base linked-structure object-reference array writers; `internal/propertyfixtures` centralizes parsed `CustomItemDatas` cryopod/custom-data builders; specialized status and malformed parser fixtures remain. |
| P3-REG-001 | `[x]` | Re-run regression tests after refactor slices. | `make verify` is the committed verification gate. |

## Phase 4: Documentation And Production Readiness

| ID | Status | Requirement | Evidence / Remaining Work |
| --- | --- | --- | --- |
| P4-DOC-001 | `[x]` | README covers install, build, CLI, library examples, scope, and mutation safety. | README is committed. |
| P4-DOC-002 | `[x]` | Supported file types and unsupported features are documented. | README and development docs. |
| P4-DOC-003 | `[x]` | Mutation APIs are documented as experimental and live-server-unverified. | README, docs, and mutation package comments. |
| P4-DEV-001 | `[x]` | Oracle regeneration, privacy rules, and safe fixture guidance are documented. | [`development.md`](development.md). |
| P4-EX-001 | `[x]` | Idiomatic Go examples exist for implemented map, player, tribe, dino, structure, equipment, local cluster, JSON, and mutation-copy workflows. | `examples/` and `examples/README.md`. |
| P4-VERIFY-001 | `[x]` | `go test ./...` passes under the repository verification target. | `make verify` runs full tests. |
| P4-VERIFY-002 | `[x]` | CLI static/local binary builds. | `make build` uses `CGO_ENABLED=0`. |
| P4-VERIFY-003 | `[x]` | CLI and example smoke tests pass on synthetic fixtures. | `cmd/arksave` and `examples` tests. |
| P4-VERIFY-004 | `[x]` | Go-only provided-data E2E smoke test is available. | `make e2e-test` exercises selected read-only APIs, stackable/dino/equipment domain JSON API/CLI export, bounded structure/base API checks, CLI commands, local tribute/cluster handling, and examples against configured private/provided data and skips without env vars. |
| P4-VERIFY-005 | `[x]` | Oracle comparison suite is rerunnable for selected implemented features. | Harness exists and records aggregate results; expanding Python/oracle coverage for every upstream example is intentionally out of scope. |
| P4-REVIEW-001 | `[blocked]` | Final production-readiness review. | Blocked until remaining Phase 2 and Phase 3 parity/refactor gaps are closed. |

## Ledger Detail Map

Use this table to jump from stable inventory IDs to the detailed ledger sections.
The detailed ledger is prose-oriented and may group several IDs under one
heading, but every inventory ID belongs to one of these ranges.

| Inventory IDs | Detailed Status Location |
| --- | --- |
| `SCOPE-*` | [`project-task-ledger.md#operating-rules`](project-task-ledger.md#operating-rules) |
| `P1-*` | [`project-task-ledger.md#phase-1-oracle-setup`](project-task-ledger.md#phase-1-oracle-setup) |
| `P2-BIN-*` | [`project-task-ledger.md#core-binary-layer`](project-task-ledger.md#core-binary-layer) |
| `P2-SAVE-*` | [`project-task-ledger.md#save-access`](project-task-ledger.md#save-access) |
| `P2-PROP-*` | [`project-task-ledger.md#property-parser`](project-task-ledger.md#property-parser) |
| `P2-OBJ-*` | [`project-task-ledger.md#object-model`](project-task-ledger.md#object-model) |
| `P2-API-*` | [`project-task-ledger.md#offline-apis`](project-task-ledger.md#offline-apis) |
| `P2-MUT-*` | [`project-task-ledger.md#mutation-apis`](project-task-ledger.md#mutation-apis) |
| `P2-EX-*` | [`project-task-ledger.md#examples`](project-task-ledger.md#examples) |
| `P3-*` | [`project-task-ledger.md#phase-3-idiomatic-go-refactor`](project-task-ledger.md#phase-3-idiomatic-go-refactor) |
| `P4-*` | [`project-task-ledger.md#phase-4-documentation-and-production-readiness`](project-task-ledger.md#phase-4-documentation-and-production-readiness) |

## Monitoring Commands

List all open, partial, or blocked inventory rows:

```sh
rg -n '^\\| [A-Z0-9-]+ \\| `\\[( |~|blocked)\\]`' docs/task-inventory.md
```

List open items across all progress docs:

```sh
rg -n "^\\s*- \\[ \\]|\\[~\\]|\\[blocked\\]" docs/project-task-ledger.md docs/phase-*.md docs/production-readiness-review.md
```
