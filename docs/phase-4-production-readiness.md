# Phase 4 Documentation And Production Readiness

Phase 4 is complete when another engineer can build, test, run, and extend the
offline parser without Python or private chat context.

For the cross-phase monitorable checklist, see
[`docs/project-task-ledger.md`](project-task-ledger.md).

## Current Status

Phase 4 is closed for documentation, Go verification, provided-data smoke
checks, CLI smoke, and production-readiness review. Phase 1, Phase 2, and Phase
3 are closed for the selected offline, fixture-backed scope. The only remaining
Phase 4 blocker is the selected oracle comparison full-suite rerun, which now
gets the oracle venv on `PATH` but still exceeds the practical interactive run
window in upstream structure parsing. Do not reopen Python oracle expansion to
work around that blockage.

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
- [x] `make verify`
- [x] Public GitHub Actions `verify` workflow runs `make verify` on `main`
      pushes, pull requests, and manual dispatches; push run `28346094077`
      passed on 2026-06-29.
- [x] `go test ./examples/...`
- [x] Public local-cluster fixture smoke for `arksave cluster`.
- [x] Public local-cluster fixture smoke for `arksave export-cluster-json`.
- [x] Optional provided-data smoke for typed local-cluster API/example coverage
      through `ARK_E2E_CLUSTER` or `ARK_E2E_CLUSTER_DIR`.
- [x] Synthetic domain JSON export tests for dinos, structures, equipment,
      stackables, players, tribes, and bases.
- [blocked] Private oracle comparison harness for selected implemented read-only Go examples;
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
      aggregate counts. The 2026-06-29 rerun did not complete as a full-suite
      run: the Make target now gives upstream subprocesses access to the oracle
      venv `python`, and a subsequent rerun progressed past the previous
      malformed-cryopod/logger failure before being interrupted while CPU-bound
      in upstream structure parsing. This is recorded as a full-suite runtime
      blockage, not a Go mismatch. A focused selected-case rerun through
      `make oracle-compare ORACLE_COMPARE_ARGS="--case tribe_list"` passed with
      the private provided save.
- [x] Expanding private oracle comparison coverage to every runnable upstream
      Python example is intentionally out of scope; use existing oracle evidence
      only when it helps selected-feature Go parity.
- [x] Final review for parser parity, API coverage, privacy, docs, and release
      readiness. Current review findings are recorded in
      `docs/production-readiness-review.md`.

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
- Remaining upstream-wide domain/API parity and complete runnable-example oracle
  coverage are not Phase 4 blockers for the selected offline scope. They remain
  documented compatibility limits until a future fixture or chosen feature makes
  them concrete requirements.
