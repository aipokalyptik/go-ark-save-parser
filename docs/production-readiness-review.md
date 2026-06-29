# Production Readiness Review

Date: 2026-06-29

Scope: Phase 4 review for the selected offline, fixture-backed Go port. This
review covers parser parity evidence, API coverage, privacy boundaries,
documentation, examples, CLI behavior, and release readiness. It does not
expand the Python oracle suite and does not inspect or commit private `.oracle`
contents.

## Result

The project is ready for user review as a production-quality offline Go base for
the implemented Ark Survival Ascended save tooling scope:

- Local `.ark`, `.arkprofile`, `.arktribe`, local cluster archive, and local
  tribute index reads are documented and covered by Go tests.
- FTP, RCON, and live server integration are intentionally unsupported.
- Mutation helpers require explicit output paths, are structurally tested, and
  remain live-server-unverified.
- The CLI and library build without Python. Python is only needed for optional
  oracle regeneration/comparison.
- Private save data, private oracle output, extracted saves, temp databases,
  debug dumps, and local snapshots are excluded from git.

## Verification Evidence

Final Phase 4 verification on 2026-06-29:

- `git diff --check && make verify`: passed. This ran `go test ./... -count=1`,
  Python script compile checks, Python script unit tests, and `CGO_ENABLED=0 go
  build -o bin/arksave ./cmd/arksave`.
- `go test ./examples/... -count=1`: passed.
- `make e2e-test` with `ARK_E2E_SAVE` set to the private provided `.ark` save
  by absolute ignored path: passed for `arkapi`, `cmd/arksave`, and `examples`.
- `make oracle-test` with `ARK_ORACLE_SAVE` and `ARK_ORACLE_TRIBUTE` set to
  private ignored paths: passed for save object enumeration and local tribute
  parsing.
- `./bin/arksave --help`: prints usage successfully.
- `git ls-remote --heads origin main`: confirmed the public `main` branch is
  present on GitHub.
- `make oracle-compare`: not completed in this run. The existing selected-case
  harness entered an upstream malformed legacy cryopod parse path and upstream
  logging attempted to spawn `python` for its hex-view helper. The active
  command path used Python 3.13 through the oracle venv, but no `python`
  executable was available for that upstream subprocess name. This is recorded
  as an oracle-environment/upstream-path blockage; no Go mismatch was produced
  by this run.

## Parity Evidence

The committed oracle comparison summary covers selected implemented aggregate
read-only cases: map/save summary, object class and object-property summaries,
local profile/tribe aggregates, player and tribe roster/relationship summaries,
dino aggregate and heatmap cases, stackable and equipment aggregate cases,
structure/base aggregate cases, domain JSON summaries, local cluster summaries,
local tribute summaries, and utility/logging behavior.

Expanding private oracle comparison coverage to every runnable upstream Python
example is intentionally out of scope. Future oracle work should be limited to
existing evidence or focused checks that validate a chosen Go feature.

## Residual Limitations

- Dynamic property parity is broad but not complete for every dedicated struct
  reader or future compound payload encoding.
- Legacy archive object parsing remains unsupported outside modern archive and
  compact local tribute index formats.
- Legacy/modded cryopod dino payloads, saddle payloads, and cosmetics remain
  fixture-gated. Modern supported `CustomItemDatas` embedded archive formats
  are parsed where implemented.
- Some upstream cryopod-location and pedigree private oracle cases remain
  blocked because upstream Python does not produce stable aggregate output on
  the supplied save.
- Deeper local cluster item/dino models and edge-specific player/tribe reports
  remain incremental until concrete local fixtures expose the required fields.
- Mutation behavior is structurally tested only. Treat generated modified saves
  as experimental until manually validated on a live server with disposable
  map data.

## Privacy And Operations

- Runtime outputs can contain private paths, IDs, class names, player/tribe
  details, locations, and upload identifiers.
- Use `--redact` for supported CLI summaries and JSON exports intended for
  logs, tickets, issue comments, or public artifacts.
- Example output is intentionally unredacted and should be treated as private
  unless run against sanitized fixtures.
- JSON and mutation outputs are written with private file permissions where the
  implementation owns file creation.

## Follow-Up Guidance

New work should stay phase-gated:

1. Do not reopen Phase 1 for broad Python oracle expansion.
2. Reopen Phase 2 only when a new Go failing test or provided-data failure
   exposes a concrete offline parser/API parity defect.
3. Treat Phase 3 as closed unless a regression proves a typed API, package, CLI,
   performance, or fixture-refactor task must be repaired.
4. For future features, add Go tests and examples first; add focused oracle
   checks only when they answer a specific chosen-feature parity question.
