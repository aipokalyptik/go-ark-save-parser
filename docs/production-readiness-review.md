# Production Readiness Review

Date: 2026-06-20

Scope: read-only subagent review of parser parity, API coverage, privacy,
documentation, examples, and production readiness. Reviewers did not inspect or
print private `.oracle` contents.

## Result

The project is a useful offline read-only foundation, but it is not yet
production-complete against the original compatibility goal. Current blockers
are parity evidence and full offline domain/API coverage, not basic build
health.

Public verification reported by reviewers:

- `go test ./...` passes.
- `make build` passes.

## Blockers

- Oracle parity evidence is still narrow. The committed oracle comparison
  summary currently covers implemented direct read-only counterparts for
  `map_summary` and `object_classes`, while Phase 4 still requires comparison
  coverage for every runnable offline Python example.
- Full offline API/domain compatibility remains incomplete. Phase 2 still has
  open work for full Player/Tribe, Dino, Structure, Equipment, Stackable, Base,
  richer local cluster models, remaining read-first object wrappers, and
  complete model-specific JSON APIs.

## High-Priority Risks

- Dynamic property parity remains incomplete. Unknown top-level property types
  and unsupported compound value encodings can still fail full object parsing.
- Legacy archive and embedded cryopod paths remain unsupported outside modern
  archive and compact tribute-index formats. This is documented, but it remains
  a blocker for broad upstream parity.
- Broad save parsing does not yet expose an upstream-style faulty-object policy.
  Current parsed-object paths can abort on the first object/property parse
  error instead of collecting faulty objects for caller policy decisions.
  Addressed after this review for save object parsing by adding
  `ParsedObjectsWithFaults`, which returns parsed objects plus per-object fault
  records while preserving the existing fail-fast `ParsedObjects` behavior.
- CLI `players` and `tribes` paths can print archive metadata while suppressing
  normalized parse failures. Automation can mistake partial output for a fully
  successful parse. Addressed after this review by returning wrapped normalized
  parse errors while preserving already-printed archive metadata.
- Local cluster uploaded dino archive parse failures are not surfaced in the
  cluster JSON model. Unsupported embedded dino formats can appear as empty or
  partially parsed uploads. Addressed after this review by recording
  per-upload `ParseError` values and exporting them as `parse_error`.

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
  archive summaries.

## Next Actions

1. Expand oracle comparison coverage one runnable offline example at a time.
2. Continue filling domain/API gaps with synthetic tests and private oracle
   comparison where runnable upstream behavior exists.
