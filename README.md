# go-ark-save-parser

Offline Go port of `ark-save-parser` for reading Ark Survival Ascended save
data from local files.

This repository is intentionally scoped to portable offline tooling:

- Supported target inputs: `.ark`, `.arkprofile`, `.arktribe`, and local cluster
  files if present in the oracle backup.
- Out of scope: FTP, RCON, and live server integration.
- Mutation APIs are experimental and live-server-unverified; they may be
  structurally tested against copied saves, but correctness inside a running
  Ark server requires manual validation outside this project.

The implementation is being built in phases:

1. Private Python oracle setup from local save data.
2. Literal Go transpilation for offline behavior parity.
3. Idiomatic Go package refactor and CLI.
4. Documentation, examples, verification, and production-readiness cleanup.

Private save data and oracle output live under `.oracle/` and are ignored by
git.
