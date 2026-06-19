# Phase 2 Literal Go Transpilation

Phase 2 is complete when the Go code mirrors upstream offline behavior closely
enough that oracle-derived tests can run against translated packages.

## Requirements

- Preserve upstream behavior first, even when the shape is not idiomatic Go.
- Keep FTP and RCON out of scope.
- Support local `.ark`, `.arkprofile`, `.arktribe`, and local cluster files where
  fixtures exist.
- Treat mutation APIs as experimental and structurally tested only.
- Add tests before implementation code for each behavior slice.
- Keep private oracle data under `.oracle/`; committed tests must use synthetic
  fixtures or sanitized expected values.

## Task Ledger

### Core Binary Layer

- [x] Add tests for byte position, bounds handling, seek/skip, peek, and
      remaining-byte behavior.
- [x] Add tests for little-endian signed/unsigned integers, floats, doubles, and
      bools.
- [x] Add tests for ARK strings: positive ASCII/null-terminated, empty strings,
      and negative UTF-16LE strings.
- [x] Add tests for UUID byte order.
- [x] Add tests for name-table lookup and missing-name behavior.
- [x] Add tests for zlib inflation wrapper behavior.
- [x] Add tests for ARK wildcard decompression edge cases.
- [x] Implement the minimal binary reader/writer to pass those tests.

### Save Access

- [x] Add synthetic SQLite `.ark` fixture tests with `custom` and `game` tables.
- [x] Port save context, header parsing, name table parsing, custom table reads,
      and object binary access.
- [x] Parse `ActorTransforms` custom table into save context.
- [x] Add private-oracle integration tests gated behind an environment variable.
- [x] Validate object enumeration against oracle data.

### Property Parser

- [x] Add tests for property terminator handling.
- [x] Add tests for primitive property types.
- [x] Add tests and parsing for raw byte and enum `ByteProperty` encodings.
- [x] Add tests for unknown struct
      fallback behavior.
- [x] Add tests and parsing for object-reference properties and simple value arrays.
- [x] Add tests and generic parsing for struct arrays.
- [x] Add tests and generic property-list parsing for struct properties.
- [x] Add tests and parsing for simple value maps and sets.
- [x] Add raw fallback preservation for packed unknown structs.
- [ ] Port remaining property parsing and declared-size realignment.
- [ ] Isolate legacy parser behavior behind explicit version/format paths.

### Object Model

- [x] Port generic game object headers and property containers.
- [x] Add read-first inventory and inventory item wrappers for parsed objects.
- [x] Add read-first object owner wrapper for parsed ownership fields.
- [x] Port actor transforms and map coordinate helpers.
- [x] Add read-first structure wrapper for parsed object fields.
- [x] Add read-first equipment wrapper for parsed inventory item fields.
- [x] Add read-first player and tribe wrappers for parsed profile containers.
- [ ] Port inventory, owner, crafter, trait, dino, structure, equipment, stackable,
      player, tribe, and local cluster data models as read-first wrappers.
- [ ] Preserve raw binary/property positions needed by mutation structural tests.

### Offline APIs

- [x] Port General API object queries.
- [x] Add archive metadata parser and local `.arkprofile` / `.arktribe` file-open wrappers.
- [x] Add first normalized tribe summary extraction surface.
- [x] Add offline CLI metadata summaries for local `.arkprofile` and `.arktribe` files.
- [x] Add first local-file Player/Tribe API surface for profile and tribe discovery/loading.
- [ ] Port full Player and Tribe APIs for parsed local files and save-contained data.
- [x] Add first read-only Structure API surface for class, owner, and location queries.
- [x] Add read-only Stackable API surface for local resource/ammo/consumable counts.
- [x] Add read-only save-info JSON export API and CLI command.
- [x] Add first read-only Equipment API surface for weapon/armor/saddle/shield queries.
- [ ] Port Dino, full Structure, full Equipment, full Stackable, Base, and full model-specific JSON APIs.
- [ ] Mark unsupported FTP/RCON examples as skipped in compatibility docs.

### Experimental Mutation

- [ ] Port copy-based modification helpers where upstream behavior can be
      translated safely.
- [ ] Require explicit output paths.
- [ ] Add structural write/reopen/reparse tests only.
- [ ] Document live-server validation as out of scope.

### Examples And Review

- [ ] Add Go examples for runnable offline Python examples.
- [ ] Compare normalized Go outputs with private Python oracle outputs where
      available.
- [ ] Run subagent spec and quality reviews on parser parity and API coverage.
