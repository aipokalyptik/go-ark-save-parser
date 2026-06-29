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

## Execution Mode

This inventory is phase-gated from this point forward. Phase 1 is closed for
the selected offline parity scope, Phase 2 is closed with documented blockers
for the selected offline, fixture-backed scope, and Phase 3 is complete for the
selected offline, fixture-backed scope. Phase 4 is closed for documentation,
verification, provided-data checks, and production-readiness review work, with
the selected-feature oracle comparison full-suite rerun blocked by upstream
runtime cost in the supplied private save. Status docs may still be updated
with implementation commits so progress stays auditable.

The accepted-scope evidence audit is [`acceptance-audit.md`](acceptance-audit.md).

## Scope Rules

| ID | Status | Requirement | Evidence / Remaining Work |
| --- | --- | --- | --- |
| SCOPE-001 | `[x]` | Offline local-file access is the compatibility target. | README scope and package docs cover `.ark`, `.arkprofile`, `.arktribe`, local cluster, and local tribute files. |
| SCOPE-002 | `[x]` | FTP is unsupported. | Documented as intentionally out of scope. |
| SCOPE-003 | `[x]` | RCON is unsupported. | Documented as intentionally out of scope. |
| SCOPE-004 | `[x]` | Cluster support is local-file-only. | `arkcluster` and CLI cluster commands read local extensionless archive files only. |
| SCOPE-005 | `[x]` | Mutation APIs are translated only as explicit copied-save workflows and marked live-server-unverified. | `arkmutation` requires output paths, has structural tests, and is documented as experimental/live-server-unverified by design. |
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
| P2-PROP-003 | `[blocked]` | Port future compound payload encodings discovered while implementing chosen offline features. | No open chosen-feature failure currently has a concrete compound payload encoding to port. Keep this blocked until a Phase 2 Go parity test or provided-data failure exposes a specific encoding. |
| P2-PROP-004 | `[blocked]` | Port legacy property/object parsing where a runnable offline oracle path exists. | Upstream uses a separate legacy binary/property parser for pre-Unreal-5.5 archives and legacy cryopod payloads. Go currently reports typed unsupported legacy errors; full port remains blocked until a committed synthetic fixture or private runnable oracle case proves exact behavior. |
| P2-OBJ-001 | `[x]` | Port generic game objects and raw position/span preservation. | `arkobject` and mutation tests cover parsing and structural spans. |
| P2-OBJ-002 | `[x]` | Port read-first wrappers for inventory, owner, structure, equipment, player, tribe, dino, base, and cluster summaries. | Implemented wrappers live under `arkobject` and `arkcluster`. |
| P2-OBJ-003 | `[x]` | Complete read-first wrappers required by the chosen offline examples. | Inventory, owner, structure, equipment, player, tribe, dino, base, crafter, and cluster wrapper surfaces used by implemented APIs/examples are covered. Add lower-priority fields only when new Go failures or fixtures expose a concrete need. |
| P2-API-001 | `[x]` | Port General API queries and fault collection. | `arkapi.GeneralAPI` tests. |
| P2-API-002 | `[x]` | Port local profile, tribe, local tribute, and local cluster parsing. | `arkprofile`, `arktribute`, `arkcluster`, and `arkapi` tests, including fault-preserving local profile/tribe/cluster/tribute batch reads and CLI directory summaries where exposed. |
| P2-API-003 | `[x]` | Port save-contained player and tribe parsing, including embedded `GameModeCustomBytes`. | `arkapi` player/tribe tests. |
| P2-API-004 | `[x]` | Port Dino, Structure, Equipment, Stackable, Base, and JSON read APIs for implemented offline workflows. | Domain tests, examples, and CLI/domain JSON tests, including provided-data E2E for stackable, dino, and equipment JSON exports plus bounded structure/base API checks. |
| P2-API-005 | `[blocked]` | Finish full dino edge behavior. | Implemented coverage includes typed pedigree tree helpers, domain JSON pedigree trees, typed malformed cryopod dino payload faults, typed unsupported cryopod saddle payload faults, and fixture-backed cryopod item transform fallback for cryopodded dino locations. Remaining legacy/modded cryopod parsing and full private pedigree comparison are blocked by malformed/unsupported cryopod payload paths or missing focused fixtures. |
| P2-API-006 | `[~]` | Finish full structure/base edge behavior. | Implemented coverage includes base/structure binary export-import round trips, typed structure health summaries, owner-location skip counters, an exact full-parse owner/location API, and shared synthetic-tested heatmap boundary semantics for in-range, out-of-range, and invalid coordinate values. Remaining upstream/private `structure_heatmap` numeric oracle parity is blocked by upstream private-save cell indexing; base customization and live-server acceptance are out of local verification scope. |
| P2-API-007 | `[blocked]` | Finish full equipment edge behavior. | Implemented coverage includes ranking candidate selection, upstream-style effective internal average stats, inventory-state summaries, upstream-list classification guards, upstream family/slot default stat tables, cursed shield defaults, inferable cursed weapon defaults, Tek sword durability, generic `CustomItemDatas` metadata, and typed unsupported cryopod saddle fault handling. Remaining exact private ranking comparison would require expanding the Python oracle suite; legacy/modded cryopod saddle payloads, cosmetics, and future default-table mismatches need concrete fixtures or failures. |
| P2-API-008 | `[blocked]` | Finish richer local cluster item/dino domain models. | Implemented coverage includes typed uploaded item/dino summaries, parse/version/component status fields, short names, embedded uploaded-dino identity and stat-level projections, JSON/CLI detail output, identity/stat aggregate counts, embedded dino level aggregates, fault-preserving directory reads, directory fault reporting, and aggregate summaries. Deeper item/dino fields are fixture-gated until chosen local-file data exposes them. |
| P2-API-009 | `[blocked]` | Finish remaining Player/Tribe edge behavior. | Implemented coverage includes typed player pawn inventory indexing, upstream-style inventory item counting, save-backed inventory summaries, path-level player inventory summaries, model-backed roster/data summaries, fault-returning roster/all summaries, player-all aggregate summaries, relationship edge counters, local death/level/experience helpers, player directory average/max level and experience summaries, tribe directory member/dino averages, and fault-preserving profile/tribe batch reads. Remaining upstream edge cases are fixture/example-gated. |
| P2-MUT-001 | `[x]` | Port copy-based DB modification, object removal, object upsert, and custom-table upsert. | `arkmutation` tests and CLI mutate commands. |
| P2-MUT-002 | `[blocked]` | Translate higher-level mutation examples where feasible. | Generated base/structure/dino/equipment binary exports round-trip through mutation imports into reopenable copied saves. Generated blueprint/base customization semantics and live-server acceptance remain explicitly unverified and outside local-only validation scope. |
| P2-EX-001 | `[x]` | Create Go equivalents for runnable offline Python examples that currently have implemented API support. | `examples/` contains committed Go examples and smoke tests. |
| P2-EX-002 | `[x]` | Compare Go example output to Python oracle output for selected implemented examples. | Forty-six aggregate cases pass; expanding Python/oracle coverage for every feasible upstream example is no longer a project goal. |

## Phase 3: Idiomatic Go Refactor

| ID | Status | Requirement | Evidence / Remaining Work |
| --- | --- | --- | --- |
| P3-PKG-001 | `[x]` | Split packages for binary, save, property, object, profile, cluster, tribute, API, mutation, and CLI surfaces. | Current package shape is documented in [`phase-3-refactor.md`](phase-3-refactor.md). |
| P3-PKG-002 | `[x]` | Further split large domain models under `arkobject` or subpackages. | Shared crafter metadata now lives in its own `arkobject` file; equipment item construction, stat calculation, default stat tables, and property coercion helpers are split; dino construction, colors, lineage, traits, and shared object-property coercion helpers are split; player, tribe, and shared profile property helpers are split; cluster item, dino, and class-name helpers are split; inventory collection, inventory item, stackable item, and shared reference helpers are split; structure construction, ownership matching, linked references, and shared scalar property helpers are split with package-shape regression tests. |
| P3-API-001 | `[x]` | Replace dynamic returns with typed Go structs, explicit errors, and fault collections for implemented paths. | Implemented typed APIs and domain JSON exports. |
| P3-API-002 | `[x]` | Replace remaining Python-shaped compatibility helpers where typed Go surfaces now exist. | `map_summary`, `parse_all`, `object_classes`, `object_summary`, `class_property_summary`, `class_lookup`, `property_filter`, and `property_positions` examples and matching CLI commands use typed `arkapi.GeneralAPI` or JSON helpers; general aggregate examples now use dedicated general path helpers instead of owning save-open/close handling; player/tribe examples use typed player path helpers where aggregate surfaces exist; `player_all`, `player_list`, `tribe_list`, `player_tribe_links`, `player_and_tribe_data`, `player_unlocked_engrams`, `local_profiles`, and the `players`/`tribes` directory CLI paths use typed roster/relation/local-profile/player-directory/tribe-directory aggregate summaries; `local_tribute`, `tribute_json`, `cluster_json`, and `cluster_typed` use typed path helpers; player inventory examples use `arkapi.PlayerInventorySummaryFromPath` and `arkapi.PlayerInventoryLookupFromPath`; equipment, dino, structure, stackable, base, heatmap, owner-location, binary export, and all-domain JSON examples now use typed `arkapi` path helpers, including shared heatmap and JSON lifecycle helpers; `equipment-best`, `equipment-rank`, `equipment-ascendant-weapon-bps`, `structure-health`, `structure-at-location`, `structure-owner-count`, `structure-owners`, `structure-heatmap`, and `dino-heatmap` CLI/example workflows use typed export/summary path helpers instead of owning save-open/write lifecycles. Remaining low-level examples are generic inspection or mutation-copy workflows where direct save/file handling is the feature. |
| P3-API-003 | `[x]` | Add remaining full typed API layers and model-specific JSON exports. | Stackable count/owner-filter path summaries, including fault-preserving aggregate CLI helper, structure tribe-ownership/owner/health/at-location path summaries plus aggregate CLI usage, structure heatmap summary, selected-structure heatmap JSON export, dino population/wild-tamable/baby/best-stat/most-mutated/wild-tamed path summaries and aggregate CLI/example usage, dino/structure/equipment/base binary export path helpers, shared JSON path constructor, general save-info/classes/parse/object/property/class path helpers, save-info/domain JSON path helpers including redacted variants, equipment inventory-state/path summary plus aggregate CLI usage, equipment best/rank path summaries plus CLI/example usage, equipment owned-by path summary plus CLI usage, equipment saddle path summary including cryopod saddles plus CLI usage, equipment path summary including cryopod saddles, player roster/player-all/tribe roster/player-tribe relation/player-and-tribe-data path summaries plus CLI/example usage, typed player profile and tribe file summary helpers with CLI usage, base aggregate/component path summaries plus base-components CLI/example usage, local tribute file/directory CLI/export lifecycles through typed path helpers, local cluster file/directory summary, cluster CLI/export lifecycles through typed path helpers, cluster JSON crafted item status, path-level dino/structure heatmap summaries, and shared heatmap summaries now have typed API helpers backed by typed lifecycle constructors; full typed dino, structure, equipment, stackable, and base collection path helpers now exist with fault preservation; typed local cluster item and dino path helpers now exist; typed player and tribe collection path helpers now exist with local fault preservation. Remaining upstream-specific player/tribe edge reports stay tracked as Phase 2 fixture-gated blockers, not open Phase 3 refactor work. |
| P3-PERF-001 | `[x]` | Add benchmarks for full save open/object enumeration, object parse, query filters, and JSON export. | Benchmarks are committed. |
| P3-PERF-002 | `[x]` | Add object cache controls and prove safe concurrency only where tested. | `arksave.Save` object row cache and concurrent raw read tests exist. |
| P3-CLI-001 | `[x]` | Implement offline CLI commands. | `inspect`, `parse`, `map-summary`, `object-classes`, `object-summary`, `property-positions`, `class-lookup`, `class-property-summary`, `property-filter`, `structure-health`, `structure-owner-count`, `structure-owners`, `structure-owner-locations`, `structure-heatmap`, `base-components`, `dinos`, `dino-wild-tamables`, `dino-babies`, `dino-best-stat`, `dino-best-base-stat`, `dino-most-mutated`, `dino-wild-tamed`, `dino-heatmap`, `equipment-summary`, `equipment-saddles`, `equipment-best`, `equipment-rank`, `equipment-ascendant-weapon-bps`, `equipment-history`, `equipment-owned-by`, `stackables`, `stackable-owned-by`, `player-inventories`, `player-roster`, `tribe-roster`, `player-tribe-links`, `players`, `tribes`, `cluster`, `cluster-summary`, `tribute`, JSON export commands, and experimental mutate commands exist. |
| P3-FIX-001 | `[x]` | Replace duplicated synthetic fixture builders with internal helpers. | `internal/testfixtures` centralizes many shared fixtures, including player pawn, inventory, structure, equipment, stackable object builders used directly by stackable API tests/benchmarks, basic dino save-object builders, dino stats/status object byte fixtures, binary `CustomItemDatas` writers, ID-table Vector struct property writers, base linked-structure object-reference array writers, player/tribe relation directory fixtures, ID-table game-object bytes with custom object-name payloads, simple ID-table int-property object rows now used directly by save-layer and arkapi general/core tests without local wrappers, and CLI archive/tribute smoke fixture calls without local wrappers; `internal/propertyfixtures` centralizes parsed `CustomItemDatas` cryopod/custom-data builders; API and object-model cryopod payload tests call shared fixtures directly. Remaining `synthetic*`/`createSynthetic*` helpers are package-local malformed payloads, purpose-built name-table headers, or domain-specific save graphs that describe the behavior under test. |
| P3-REG-001 | `[x]` | Re-run regression tests after refactor slices. | `make verify` is the committed verification gate. |

## Phase 4: Documentation And Production Readiness

| ID | Status | Requirement | Evidence / Remaining Work |
| --- | --- | --- | --- |
| P4-DOC-001 | `[x]` | README covers install, build, CLI, library examples, scope, and mutation safety. | README is committed and Phase 4 reviewed. |
| P4-DOC-002 | `[x]` | Supported file types and unsupported features are documented. | README and development docs cover local `.ark`, `.arkprofile`, `.arktribe`, local cluster, local tribute, and unsupported FTP/RCON/live-server features. |
| P4-DOC-003 | `[x]` | Mutation APIs are documented as experimental and live-server-unverified. | README, docs, package comments, and Phase 4 review document copied-save-only mutation safety. |
| P4-DOC-004 | `[x]` | Public Go packages have package-level documentation for library consumers. | `doc.go` files exist for the public parser, API, object, archive, profile, cluster, tribute, mutation, and logging packages; `go doc ./arkapi`, `go doc ./arksave`, and `go doc ./arkmutation` were verified on 2026-06-29. |
| P4-DEV-001 | `[x]` | Oracle regeneration, privacy rules, and safe fixture guidance are documented. | [`development.md`](development.md) and Phase 4 review cover regeneration and privacy boundaries. |
| P4-EX-001 | `[x]` | Idiomatic Go examples exist for implemented map, player, tribe, dino, structure, equipment, local cluster, JSON, and mutation-copy workflows. | `examples/` and `examples/README.md`; `go test ./examples/... -count=1` passed on 2026-06-29. |
| P4-VERIFY-001 | `[x]` | `go test ./...` and `go vet ./...` pass under the repository verification target. | `git diff --check && make verify` passed on 2026-06-29. |
| P4-VERIFY-002 | `[x]` | CLI static/local binary builds. | `make verify` ran `CGO_ENABLED=0 go build -o bin/arksave ./cmd/arksave` successfully on 2026-06-29. |
| P4-VERIFY-003 | `[x]` | CLI and example smoke tests pass on synthetic fixtures. | `make verify`, `go test ./examples/... -count=1`, and the built-in CLI usage smoke passed on 2026-06-29. |
| P4-VERIFY-004 | `[x]` | Go-only provided-data E2E smoke test is available. | `make e2e-test` passed on 2026-06-29 with `ARK_E2E_SAVE` and `ARK_E2E_SAVE_DIR` set to private provided Valguero paths under ignored `.oracle`, covering `arkapi`, `cmd/arksave`, and `examples`. |
| P4-VERIFY-005 | `[blocked]` | Oracle comparison suite is rerunnable for selected implemented features. | Existing committed summary remains reference evidence. The Make target now gives upstream subprocesses access to the oracle venv `python`, and a focused `tribe_list` rerun passed through `make oracle-compare ORACLE_COMPARE_ARGS="--case tribe_list"` with the private provided save. The 2026-06-29 full-suite rerun progressed past the previous malformed-cryopod/logger failure before being interrupted while CPU-bound in upstream structure parsing. Expanding Python/oracle coverage remains out of scope. |
| P4-VERIFY-006 | `[x]` | CLI release binary exposes build metadata. | `make build` stamps version, commit, and UTC build time through ldflags; `./bin/arksave --version` was smoke-tested on 2026-06-29. |
| P4-VERIFY-007 | `[x]` | Public GitHub Actions verification runs on push, pull request, and manual dispatch. | `.github/workflows/verify.yml` runs `make verify` on `main` pushes, pull requests, and explicit manual dispatches without private oracle data; the README badge links to current status and historical passing runs include `28346712289` and `28347063999`. |
| P4-REVIEW-001 | `[x]` | Final production-readiness review. | [`production-readiness-review.md`](production-readiness-review.md) refreshed on 2026-06-29 with verification evidence and residual limitations. |

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

Use the portable status target to list the remaining open, partial, or blocked
rows without depending on `rg`:

```sh
make status
```

List all open, partial, or blocked inventory rows:

```sh
rg -n '^\\| [A-Z0-9-]+ \\| `\\[( |~|blocked)\\]`' docs/task-inventory.md
```

List open items across all progress docs:

```sh
rg -n "^\\s*- \\[ \\]|\\[~\\]|\\[blocked\\]" docs/project-task-ledger.md docs/phase-*.md docs/production-readiness-review.md
```
