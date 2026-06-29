# Acceptance Audit

Date: 2026-06-29

This point-in-time audit maps the original offline-only Go port acceptance
criteria to repository evidence. It is intentionally scoped to local-file
offline access. FTP, RCON, and live-server mutation validation are outside the
accepted scope.

## Repository And Phase Evidence

| Requirement | Status | Evidence |
| --- | --- | --- |
| Public GitHub repository exists. | Met | `gh repo view aipokalyptik/go-ark-save-parser` reports `visibility=PUBLIC`, `isPrivate=false`, URL `https://github.com/aipokalyptik/go-ark-save-parser`. |
| `main` is the pushed working branch. | Met | `origin` is `https://github.com/aipokalyptik/go-ark-save-parser.git`; the default branch is `main`; pushed `main` is verified with `git ls-remote --heads origin main`. |
| Phase 1 oracle setup is recorded. | Met | [`task-inventory.md`](task-inventory.md) rows `P1-001` through `P1-010`; [`phase-1-report.md`](phase-1-report.md); [`oracle-summary.md`](oracle-summary.md). |
| Phase 2 offline Go transpilation is implemented for selected runnable offline scope. | Met with documented blockers | [`task-inventory.md`](task-inventory.md) Phase 2 rows. Remaining rows are explicitly fixture-gated, upstream-blocked, live-server-unverified, or outside the accepted local-only scope. |
| Phase 3 idiomatic Go package/API/CLI refactor is closed. | Met | [`phase-3-refactor.md`](phase-3-refactor.md) and `P3-*` rows in [`task-inventory.md`](task-inventory.md). |
| Phase 4 documentation and production-readiness work is closed for accepted scope. | Met with documented blockers | [`phase-4-production-readiness.md`](phase-4-production-readiness.md), [`production-readiness-review.md`](production-readiness-review.md), and `P4-*` rows in [`task-inventory.md`](task-inventory.md). |

## Scope Criteria

| Criterion | Status | Evidence |
| --- | --- | --- |
| Offline local-file access is the compatibility target. | Met | README and package docs cover `.ark`, `.arkprofile`, `.arktribe`, local cluster, and local tribute files. |
| FTP support is intentionally omitted. | Met | `SCOPE-002` and CLI unsupported-command tests document/cover this. |
| RCON support is intentionally omitted. | Met | `SCOPE-003` and CLI unsupported-command tests document/cover this. |
| Cluster support is local-file-only. | Met | `arkcluster` and CLI cluster commands operate on local extensionless archive files. |
| Mutation APIs are copied-save workflows only. | Met | `arkmutation` requires explicit output paths, has structural tests, and docs/CLI output mark mutation copies experimental and live-server-unverified. |
| Private save data and raw oracle output are not committed. | Met | `.gitignore`, development docs, and sanitized oracle docs define ignored `.oracle`, extracted saves, snapshots, debug dumps, temp DBs, and private generated output. |
| Do not expand Python oracle coverage broadly. | Met | Current docs preserve existing oracle evidence and direct future work toward Go tests/examples unless a focused parity question requires a narrow oracle check. |

## Verification Criteria

| Gate | Status | Evidence |
| --- | --- | --- |
| `go test ./...` passes. | Met | `make verify` passed locally on 2026-06-29 and includes `go test ./... -count=1`. |
| `go vet ./...` passes. | Met | `make verify` passed locally on 2026-06-29 and includes `go vet ./...`. |
| Static/local CLI builds without system SQLite. | Met | `make verify` passed locally on 2026-06-29 and includes `CGO_ENABLED=0 go build -trimpath -o bin/arksave ./cmd/arksave`. |
| Public CI verifies the repository without private data. | Met | `.github/workflows/verify.yml` runs `make verify` on `main` pushes, pull requests, and manual dispatch. The README badge links to the current workflow status; historical passing runs include `28346712289` and `28347063999`. |
| Provided-data Go E2E runs against private local save data. | Met | `make e2e-test` passed on 2026-06-29 with `ARK_E2E_SAVE` and `ARK_E2E_SAVE_DIR` pointing at ignored private Valguero save paths, covering `arkapi`, `cmd/arksave`, and `examples`. |
| Selected Python oracle comparison remains reproducible where practical. | Met with runtime blocker | Focused `make oracle-compare ORACLE_COMPARE_ARGS="--case tribe_list"` passed. The full selected suite rerun is blocked by upstream runtime cost in structure parsing, not by a Go mismatch. |

## Remaining Documented Blockers

These blockers do not contradict the accepted offline-only scope, but they are
not proven complete:

- Future compound property encodings without a concrete failing fixture.
- Legacy property/object parsing outside modern archives and compact local
  tribute index files.
- Legacy/modded cryopod dino/saddle payloads, cosmetics, and full private
  pedigree/cryopod-location parity where upstream does not produce stable
  aggregate output on the supplied private save.
- Full structure heatmap oracle comparison where upstream private-save cell
  indexing fails.
- Deeper local cluster item/dino fields and long-tail player/tribe/equipment
  edge cases until concrete local fixtures expose them.
- Higher-level semantic mutation workflows and live-server acceptance.

## Conclusion

The repository satisfies the accepted production-readiness criteria for a
portable, statically built, offline-only Go base for Ark Survival Ascended save
tooling. Remaining work is explicitly documented as fixture-gated,
upstream-blocked, or live-server-unverified.
