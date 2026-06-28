# Phase 4 Documentation And Production Readiness

Phase 4 is complete when another engineer can build, test, run, and extend the
offline parser without Python or private chat context.

For the cross-phase monitorable checklist, see
[`docs/project-task-ledger.md`](project-task-ledger.md).

## Current Status

Phase 4 is not the active execution phase. The documentation and verification
items below record useful ahead-of-phase work, but final cleanup and production
readiness review must wait until Phase 3 is closed. Keep future documentation
edits narrow while Phase 3 is active, limited to status tracking and direct
support for Go refactor work.

## Documentation Tasks

- [~] Write README with build, CLI, library, example, scope, and mutation safety
      notes.
- [~] Document supported file types: `.ark`, `.arkprofile`, `.arktribe`,
      extensionless local cluster files, and local tribute index files.
- [~] Document unsupported features: FTP, RCON, live server integration, and
      legacy archive object parsing outside modern archive/local tribute index
      formats.
- [~] Document mutation APIs as experimental and live-server-unverified.
- [~] Document oracle regeneration using `~/Downloads/SavedArks.tar.bz2`.
- [~] Document privacy rules and ignored paths.
- [~] Document runtime output sensitivity and write JSON/mutation outputs with
      private file permissions.
- [~] Add and document opt-in CLI redaction for aggregate summaries and JSON
      exports.
- [~] Add basic large-file guards for whole-file profile, tribe, and local
      cluster archive reads.
- [~] Document how to add new oracle fixtures safely.
- [~] Add standalone Go examples for implemented offline workflows.

## Verification Tasks

- [~] `go test ./...`
- [~] `make build`
- [~] `make verify`
- [~] `go test ./examples/...`
- [~] Public local-cluster fixture smoke for `arksave cluster`.
- [~] Public local-cluster fixture smoke for `arksave export-cluster-json`.
- [~] Optional provided-data smoke for typed local-cluster API/example coverage
      through `ARK_E2E_CLUSTER` or `ARK_E2E_CLUSTER_DIR`.
- [~] Synthetic domain JSON export tests for dinos, structures, equipment,
      stackables, players, tribes, and bases.
- [~] Private oracle comparison harness for selected implemented read-only Go examples;
      current aggregate status is in `docs/oracle-comparison-summary.md` and
      covers map summary, object classes, save-info JSON export, local
      profile/tribe aggregate counts, local player deaths and unlocked engram
      set aggregates, player and tribe list aggregates, save-level player and
      tribe relation aggregates,
      dino aggregate counts, no-cryopod dino best-stat selection,
      no-cryopod dino heatmap cell aggregates,
      property-position metadata aggregates, property-filtered object counts,
      player inventory item/location
      aggregates, stackable item/quantity counts, equipment longneck blueprint
      damage aggregates, equipment best weapon/armor value aggregates,
      canonical direct equipment class counts, stable equipment ranking
      aggregates, dino domain JSON aggregate counts,
      structure owner and map-coordinate/connected counts, local cluster JSON
      aggregate counts, local tribute aggregate counts, and tribute JSON
      aggregate counts.
- [~] Expanding private oracle comparison coverage to every runnable upstream
      Python example is intentionally out of scope; use existing oracle evidence
      only when it helps selected-feature Go parity.
- [ ] Final review for parser parity, API coverage, privacy, docs, and release
      readiness. Current ahead-of-phase review findings are recorded in
      `docs/production-readiness-review.md`; final production readiness still
      waits for Phase 3 closure and the remaining gaps below.

## Remaining Production Gaps

- Full domain parity for legacy/modded cryopod variants, legacy/modded saddle
  payloads and cosmetics inside cryopods, richer pedigree rendering/tree export
  helpers beyond JSON child/descendant references, richer equipment stats,
  player/tribe details, and base import/export remains incomplete. Modern
  cryopod dino/status and saddle payloads can be parsed when `CustomItemDatas`
  uses the supported embedded archive formats.
- Upstream cryopod-location and pedigree examples remain blocked as private
  oracle cases because upstream Python currently hits malformed embedded
  cryopod parsing on the supplied save before stable aggregate output.
- Legacy archive object parsing remains unsupported outside modern archive and
  compact local tribute index formats.
- Mutation helpers are structurally tested only and require live-server manual
  validation before being treated as production-safe for real servers.
- Latest read-only review found remaining production blockers in full
  domain/API parity and complete runnable-example oracle coverage. Runtime
  redaction now exists for supported CLI summaries and JSON exports,
  fault-tolerant domain scans exist for the currently implemented object-scan
  APIs, and whole-file profile/tribe/cluster reads have default size guards.
  Future commands still need explicit privacy and resource-bound review before
  committed output is considered safe.
