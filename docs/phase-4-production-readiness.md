# Phase 4 Documentation And Production Readiness

Phase 4 is complete when another engineer can build, test, run, and extend the
offline parser without Python or private chat context.

## Documentation Tasks

- [x] Write README with build, CLI, library, example, scope, and mutation safety
      notes.
- [x] Document supported file types: `.ark`, `.arkprofile`, `.arktribe`, and
      extensionless local cluster files.
- [x] Document unsupported features: FTP, RCON, live server integration, and
      legacy `.arktributetribe` parsing.
- [x] Document mutation APIs as experimental and live-server-unverified.
- [x] Document oracle regeneration using `~/Downloads/SavedArks.tar.bz2`.
- [x] Document privacy rules and ignored paths.
- [x] Document runtime output sensitivity and write JSON/mutation outputs with
      private file permissions.
- [x] Document how to add new oracle fixtures safely.
- [x] Add standalone Go examples for implemented offline workflows.

## Verification Tasks

- [x] `go test ./...`
- [x] `make build`
- [x] `go test ./examples/...`
- [x] Public local-cluster fixture smoke for `arksave cluster`.
- [x] Public local-cluster fixture smoke for `arksave export-cluster-json`.
- [x] Synthetic domain JSON export tests for dinos, structures, equipment,
      stackables, and bases.
- [ ] Private oracle comparison suite for every runnable Python example.
- [ ] Final review for parser parity, API coverage, privacy, docs, and release
      readiness.

## Remaining Production Gaps

- Full domain parity for dino stats, cryopods, pedigrees, richer equipment
  stats, player/tribe details, and base import/export remains incomplete.
- Legacy `.arktributetribe` local tribute archives remain unsupported.
- Mutation helpers are structurally tested only and require live-server manual
  validation before being treated as production-safe for real servers.
- Latest read-only review found remaining production blockers in oracle parity
  evidence, runtime redaction modes, large-file hardening, and full domain/API
  parity. Those are not release-blocking for continued port work, but they keep
  final production readiness open.
