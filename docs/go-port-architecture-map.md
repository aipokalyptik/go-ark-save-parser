# Go Port Architecture Map

This map translates upstream `ark-save-parser` responsibilities into the
offline-only Go package structure. It is based on upstream commit
`4f7cc91fb96a080321bfbc884ba81bd897f72c49`.

## Build Order

1. `arkbinary`: primitive binary reader/writer, string/name-table handling,
   UUID handling, zlib inflation, and ARK wildcard decompression.
2. `arksave`: SQLite-backed `.ark` access, custom tables, save headers, save
   context, actor transforms, class lookup, object binary access.
3. `arkproperty`: dynamic property parsing, arrays, maps, sets, known structs,
   unknown struct recovery, and legacy parser isolation.
4. `arkobject`: generic game object plus domain wrappers.
5. `arkapi`: offline high-level query APIs.
6. `arkcluster`: local-file cluster parsing only if local fixtures are present.
7. `cmd/arksave`: offline CLI.
8. `arkmutation`: experimental copy-based mutation helpers, structurally tested
   but live-server-unverified.

## Upstream To Go Mapping

| Upstream area | Responsibility | Go package | Main risks |
|---|---|---|---|
| `saves/asa_save.py` | Public save facade and initialization | `arksave` | Keep offline constructors only; separate read-only and mutation-capable stores. |
| `saves/save_connection.py` | SQLite `game`/`custom` access, headers, class lookup | `arksave` | UUID bytes are big-endian while scalar fields are little-endian. |
| `saves/save_context.py` | Names, versions, map, sections, actor transforms | `arksave` | Avoid global mutable state; pass context explicitly. |
| `_base_value_parser.py`, `_binary_reader_base.py` | Primitive reads, validation, names, peeking | `arkbinary` | Positive strings are null-terminated byte strings; negative lengths are UTF-16LE. |
| `ark_binary_parser.py` | High-level parser, wildcard decompression, actor transforms | `arkbinary`, `arksave` | Wildcard decompression must be byte-for-byte tested. |
| `ark_property.py`, `array_property.py`, `ark_set.py` | Dynamic property parsing | `arkproperty` | Highest-risk area: `None` terminators, declared sizes, unknown structs, UE5.5 quirks. |
| `_legacy_parsing/*` | Legacy parser fallback | `internal/legacy` | Isolate behind format/version switches. |
| `ark_object.py`, `ark_archive.py` | Object/archive parsing and cluster object paths | `arkobject`, `arkcluster` | Archive objects and save DB objects have different header paths. |
| `object_model/ark_game_object.py` | Generic object with blueprint, section, location, properties | `arkobject` | Preserve raw positions for mutation structural tests. |
| `object_model/misc/*` | Inventory, owners, crafters, traits | `arkobject` | Many upstream methods mutate binaries; gate write helpers. |
| `object_model/dinos/*` | Dino, tamed/wild/baby/stats/pedigree/cryopod | `arkobject/dino` | Cryopods embed archive data; stats and ownership need oracle tests. |
| `object_model/structures/*` | Structures, inventory structures, ownership | `arkobject/structure` | Blueprint heuristics and missing inventory references. |
| `object_model/equipment/*`, `stackables/*` | Equipment, weapons, armor, saddles, resources | `arkobject/item` | Large class constants should be generated or data-driven. |
| `player/*`, `ark_tribe.py` | Player and tribe archives | `arkobject/player`, `arkobject/tribe` | Data may live in local files or save custom bytes. |
| `object_model/cluster_data/*` | Local cluster item/dino data | `arkcluster` | Local-file support only; no network assumptions. |
| `api/general_api.py` | Generic object query facade | `arkapi` | Prefer explicit query options over mutable default config. |
| `api/dino_api.py`, `structure_api.py`, `equipment_api.py`, `stackable_api.py`, `player_api.py`, `base_api.py` | Domain query APIs | `arkapi` | Split read APIs from experimental mutation APIs. |
| `api/json_api.py`, `utils/json_utils.py` | JSON export | `arkapi` or `arkexport` | Build after stable model structs exist. |
| `ftp/*`, `api/rcon_api.py` | Network acquisition/control | Not ported | Explicitly out of scope. |

## Translation Risks

- Name-table and string handling are correctness-critical.
- Python allows dynamic property values and inheritance-heavy wrappers; Go needs
  explicit typed containers plus safe fallback values for unknown properties.
- Parser state must be instance-local to keep future goroutine parsing race-free.
- Mutation helpers must require explicit output paths and must never modify a
  source save in place by default.
- Live-server validity of modified saves is outside the automated test boundary.
