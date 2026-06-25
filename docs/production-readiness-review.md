# Production Readiness Review

Date: 2026-06-20

Scope: read-only subagent review of parser parity, API coverage, privacy,
documentation, examples, and production readiness. Reviewers did not inspect or
print private `.oracle` contents.

## Result

The project is a useful offline read-only foundation, but it is not yet
production-complete against the original compatibility goal. Current blockers
are complete runnable-example oracle coverage and full offline domain/API
coverage, not basic build health.

Public verification reported by reviewers:

- `go test ./...` passes.
- `make build` passes.

## Blockers

- Oracle parity evidence has expanded but is still incomplete. The committed
  oracle comparison summary currently covers thirty-eight implemented aggregate
  read-only cases: `map_summary`, `object_classes`, `object_summary`,
  `export_json`, `class_lookup`, `class_property_summary`, `local_profiles`,
  `local_profile_player_aggregates`,
  `player_unlocked_engrams`, `player_all`, `player_tribe_links`,
  `player_inventory`, `dino_filter`,
  `dino_best_stat_no_cryos`, `dino_best_base_stat`, `dino_most_mutated`,
  `dino_babies`, `dino_wild_tamables`, `dino_wild_tamed`, `property_filter`,
  `stackable_count`, `stackable_owned_by`, `domain_json_stackables`,
  `equipment_longneck_blueprint_damage`, `equipment_best`,
  `equipment_summary`, `equipment_rank`, `equipment_ascendant_weapon_bps`,
  `equipment_saddles`,
  `equipment_owned_by`, `structure_owner_count`, `structure_owners`,
  `structure_at_location`, `base_components`, `domain_json_dinos`,
  `cluster_json`, `local_tribute`, and `tribute_json`.
  Phase 4 still requires comparison
  coverage for every runnable offline Python example where a Go counterpart
  exists or is feasible.
- Full offline API/domain compatibility remains incomplete. Phase 2 still has
  open work for full Player/Tribe, Dino, Structure, Equipment, Stackable, Base,
  richer local cluster models, remaining read-first object wrappers, and
  complete model-specific JSON APIs.

## High-Priority Risks

- Dynamic property parity remains incomplete, but several previously high-risk
  compound encodings are now covered: enum-keyed maps, struct-keyed map body
  skipping with reader realignment, unsupported map value/set element body
  skipping, raw preservation for unknown top-level property payloads, and
  fixed-layout `Rotator`, `Quat`, `Color`, and `LinearColor` structs. Remaining
  unsupported compound value encodings and dedicated struct readers can still
  fail full object parsing.
- Legacy archive and legacy/modded embedded cryopod paths remain unsupported
  outside modern archive and compact tribute-index formats. Modern cryopod
  dino/status payloads now have a parsed API path, but broad upstream parity
  still requires legacy/modded variants plus saddle/cosmetic payload support.
- Broad save parsing now exposes an initial upstream-style faulty-object policy
  through `arksave.ParsedObjectsWithFaults` and `arkapi.GeneralAPI.ObjectsWithFaults`.
  Several domain APIs also expose partial-success scans, including dino,
  structure, equipment, stackable, base, and save-contained player and tribe
  parsing.
  Remaining domain surfaces still need to adopt this pattern where full-object
  parsing can encounter unsupported property encodings.
  Addressed after this review for save object parsing by adding
  `ParsedObjectsWithFaults`, which returns parsed objects plus per-object fault
  records while preserving the existing fail-fast `ParsedObjects` behavior.
  Addressed further for the implemented dino, structure, equipment, stackable,
  and base APIs by adding `AllWithFaults` variants for partial-success scans.
- CLI `players` and `tribes` paths can print archive metadata while suppressing
  normalized parse failures. Automation can mistake partial output for a fully
  successful parse. Addressed after this review by returning wrapped normalized
  parse errors while preserving already-printed archive metadata.
- Local cluster uploaded dino archive parse failures are not surfaced in the
  cluster JSON model. Unsupported embedded dino formats can appear as empty or
  partially parsed uploads. Addressed after this review by recording
  per-upload `ParseError` values, exporting them as `parse_error`, and showing
  them in cluster CLI summaries.
- Save-contained player/tribe parity is no longer blocked on embedded
  `GameModeCustomBytes` support. Save-backed player and tribe APIs now scan
  game-table `PrimalPlayerDataBP`/`PrimalTribeData` objects and embedded
  `GameModeCustomBytes` player/tribe archives, merge records by stable IDs,
  and preserve partial-parse fault reporting through the `WithFaults` methods.

## Medium-Priority Risks

- Best-effort archive parsing records per-object property errors, but some
  higher-level callers can still treat partial data as authoritative unless
  strict modes or explicit parse-status fields are used. Addressed after this
  review for archive/profile/tribe readers by adding archive property-error
  summaries and profile/tribe convenience status accessors.
- Mutation helpers are structurally tested only. This is correctly documented as
  live-server-unverified, but upstream behavioral parity is not proven.
- Example privacy guidance is weaker than CLI privacy guidance. Example outputs
  can contain paths, IDs, class names, player/tribe details, locations, and
  upload identifiers, but the examples README does not repeat the privacy
  warning or provide redaction-mode equivalents. Addressed after this review by
  documenting that example output is unredacted and pointing users to
  `arksave --redact` for safer aggregate output.
- Legacy/unsupported archive behavior is documented, but user-facing CLI/API
  error behavior should be made more explicit and tested. Partially addressed
  after this review by printing aggregate property parse-error counts in CLI
  archive summaries. Further addressed by making `arksave parse` perform a
  fault-tolerant full-object parse summary with parsed-object and parse-fault
  counts; on the large private Valguero save this remains too slow for the
  default oracle comparison suite and should be treated as a manual smoke path.
- Some upstream read-only examples are not stable oracle candidates on the
  private save yet. In particular, `DinoApi.get_all_babies(include_wild=True)`
  produced high-volume parser debug output and previously failed in an embedded
  cryopod path before returning a stable aggregate; defer that comparison until
  the remaining legacy/modded cryopod handling is improved or a quieter
  upstream invocation is available.

## Next Actions

1. Expand oracle comparison coverage one runnable offline example at a time.
2. Continue filling domain/API gaps with synthetic tests and private oracle
   comparison where runnable upstream behavior exists.
3. Continue routing high-volume CLI/examples through fault-tolerant object scans
   where upstream behavior skips faulty rows instead of aborting the run.
