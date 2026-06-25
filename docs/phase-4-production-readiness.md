# Phase 4 Documentation And Production Readiness

Phase 4 is complete when another engineer can build, test, run, and extend the
offline parser without Python or private chat context.

## Documentation Tasks

- [x] Write README with build, CLI, library, example, scope, and mutation safety
      notes.
- [x] Document supported file types: `.ark`, `.arkprofile`, `.arktribe`,
      extensionless local cluster files, and local tribute index files.
- [x] Document unsupported features: FTP, RCON, live server integration, and
      legacy archive object parsing outside modern archive/local tribute index
      formats.
- [x] Document mutation APIs as experimental and live-server-unverified.
- [x] Document oracle regeneration using `~/Downloads/SavedArks.tar.bz2`.
- [x] Document privacy rules and ignored paths.
- [x] Document runtime output sensitivity and write JSON/mutation outputs with
      private file permissions.
- [x] Add and document opt-in CLI redaction for aggregate summaries and JSON
      exports.
- [x] Add basic large-file guards for whole-file profile, tribe, and local
      cluster archive reads.
- [x] Document how to add new oracle fixtures safely.
- [x] Add standalone Go examples for implemented offline workflows.

## Verification Tasks

- [x] `go test ./...`
- [x] `make build`
- [x] `go test ./examples/...`
- [x] Public local-cluster fixture smoke for `arksave cluster`.
- [x] Public local-cluster fixture smoke for `arksave export-cluster-json`.
- [x] Synthetic domain JSON export tests for dinos, structures, equipment,
      stackables, players, tribes, and bases.
- [x] Private oracle comparison harness for implemented read-only Go examples;
      current aggregate status is in `docs/oracle-comparison-summary.md` and
      covers map summary, object classes, save-info JSON export, local
      profile/tribe aggregate counts, local player deaths and unlocked engram
      set aggregates, player list aggregates, save-level player and tribe
      relation aggregates,
      dino aggregate counts, no-cryopod dino best-stat selection,
      property-filtered object counts, player inventory item/location
      aggregates, stackable item/quantity counts, equipment longneck blueprint
      damage aggregates, equipment best weapon/armor value aggregates,
      canonical direct equipment class counts, stable equipment ranking
      aggregates, dino domain JSON aggregate counts,
      structure owner and map-coordinate/connected counts, local cluster JSON
      aggregate counts, local tribute aggregate counts, and tribute JSON
      aggregate counts.
- [ ] Private oracle comparison suite for every runnable Python example.
- [x] Final review for parser parity, API coverage, privacy, docs, and release
      readiness. Current review findings are recorded in
      `docs/production-readiness-review.md`; production readiness is still
      blocked by the remaining gaps below.

## Remaining Production Gaps

- Full domain parity for legacy/modded cryopod variants, legacy/modded saddle
  payloads and cosmetics inside cryopods, richer pedigree rendering/tree export
  helpers beyond JSON child/descendant references, richer equipment stats,
  player/tribe details, and base import/export remains incomplete. Modern
  cryopod dino/status and saddle payloads can be parsed when `CustomItemDatas`
  uses the supported embedded archive formats.
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
