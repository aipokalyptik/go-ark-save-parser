# Upstream Oracle Classification

Upstream reference: `VincentHenauGithub/ark-save-parser` commit
`4f7cc91fb96a080321bfbc884ba81bd897f72c49`.

This classification is scoped to the offline Go port. FTP and RCON behavior is
intentionally excluded. Mutation examples are copy-only structural candidates
and remain live-server-unverified.

## Packaged Tests

The upstream `tests/` suite is not self-sufficient in a public checkout. It
expects private fixture data under `tests/test_data`, including map folders such
as `set_1/Ragnarok_WP/Ragnarok_WP.ark`, `set_2/Ragnarok_WP/Ragnarok_WP.ark`,
base export fixtures, and cluster data. Without that corpus, the suite errors
before exercising parser behavior.

Blocked areas:

- `test_00_basic_parsing.py`: fixed Ragnarok save expectations.
- `test_dino_api.py`: fixed dino counts plus mutation-copy checks.
- `test_equipment_api.py`: fixed equipment counts plus generated item reparse
  checks.
- `test_structure_api.py`: fixed structure counts.
- `test_player_api.py`: fixed player/tribe/pawn counts and cluster fixtures.
- `test_base_api.py`: fixed base fixture imports/exports and mutation-copy
  checks.
- `test_json_api.py`: fixed map JSON export expectations.
- `test_asa_save.py`: fixed object/class counts and actor transform mutation-copy
  checks.

## Arbitrary-Save Testbench

The upstream `testbench/` directory is the primary Phase 1 oracle target because
it accepts any local `.ark` save dropped under `testbench/test_save`.

Runnable against supplied private map saves:

- `testbench/tests/test_01_parsing.py`: load, faulty-object policy, game time,
  object counts, `GeneralApi`.
- `testbench/tests/test_02_dinos.py`: dino totals, wild/tamed/cryopod/baby
  breakdowns, baby-stage invariants.
- `testbench/tests/test_03_structures.py`: structure counts and inventory
  structures.
- `testbench/tests/test_04_equipment.py`: armor, weapons, saddles, shields.
- `testbench/tests/test_05_players.py`: players, tribes, pawns, pawn inventory
  resolution.

Side effects remain private and ignored:

- `testbench/snapshots/`
- `testbench/debug_dumps/`

## Offline Read-Only Examples

These examples can be used as oracle cases after replacing FTP downloads or
hard-coded placeholder paths with local `.ark` paths:

- `examples/basic_parsing/ex_00_parse_all.py`
- `examples/basic_parsing/ex_01_get_obj_by_uuid.py`
- `examples/basic_parsing/ex_02_get_obj_by_class_string.py`
- `examples/basic_parsing/ex_03_reading_property_definitions_and_positions.py`
  is covered by `examples/property_positions`, which compares property name
  offsets, value offsets, encoded byte spans, and property index-position
  aggregates without committing property names or raw bytes.
- `examples/basic_parsing/ex_05_get_objects_with_property.py`
- `examples/basic_parsing/ex_06_find_all_properties_for_objects.py`
- `examples/dino_api/ex_01_generic_filter.py`
- `examples/dino_api/ex_03_get_dino_with_highest_stat.py`
- `examples/dino_api/ex_04_get_tamed_rock_drake_with_highest_base_stam.py`
- `examples/dino_api/ex_05_get_dino_with_most_mutations.py`
- `examples/dino_api/ex_07_get_all_babies.py`
- `examples/dino_api/ex_08_get_all_wild_tamables.py`
- `examples/dino_api/ex_09_is_wild_tamed_dino.py`
- `examples/dino_api/ex_10_get_cryopod_location.py`
- `examples/dino_api/ex_11_get_dino_pedigrees.py`
- `examples/equipment_api/ex_01_get_armor_with_highest_durability.py`
- `examples/equipment_api/ex_02_get_highest_dmg_weapon.py`
- `examples/equipment_api/ex_03_get_highest_dmg_longneck_blueprint.py`
- `examples/equipment_api/ex_04_get_all_ascended_weapon_bps.py`
- `examples/equipment_api/ex_05_get_owner_of_items.py`
- `examples/equipment_api/ex_06_rank_all_equipment.py`
- `examples/equipment_api/ex_08_get_all_saddles.py`
- `examples/player_api/ex_01a_get_all_players.py`
- `examples/player_api/ex_01c_get_all_player_and_tribe_data.py`
- `examples/player_api/ex_05_get_unlocked_engrams.py`
- `examples/stackable_api/ex_01_get_nr_of_resource.py`
- `examples/stackable_api/ex_02_get_nr_of_arb_owned_by.py`
- `examples/structure_api/ex_01_get_all_vaults_owned_by.py`
- `examples/structure_api/ex_02_get_all_structures_at_location.py`
- `examples/structure_api/ex_04_get_nr_of_structures_owned_by_tribe.py`
- `examples/structure_api/ex_06_print_struycture_owner.py`
- `examples/other/ex_01_logging_configuration.py` is covered by
  `examples/logging_config` as a standalone logging-helper demonstration. The
  Go version keeps logging state process-local and does not persist global
  config files.

## Export And Heatmap Examples

These are offline-compatible but produce files or images, so outputs stay under
`.oracle/output`:

- `examples/base_api/ex_01_export_base_from_save.py`
- `examples/base_api/ex_03_get_all_connected_sets_of_structures.py`
- `examples/dino_api/ex_02a_dino_by_stat_heatmap.py` is implemented as the
  lower-level `DinoAPI.Heatmap` helper plus the `examples/dino_heatmap` JSON
  summary example. The stat-filtered printed dino list is covered separately by
  `examples/dino_best_stat` and dino filter examples.
- `examples/dino_api/ex_02b_tamed_dino_heatmap.py` is implemented as the same
  `DinoAPI.Heatmap` helper plus `examples/dino_heatmap`; owner/name printouts
  remain intentionally excluded from committed example output.
- `examples/structure_api/ex_03_create_structure_heatmap.py` is implemented as
  `examples/structure_heatmap`, which writes a compact JSON summary instead of
  an image. Private oracle comparison is currently blocked because upstream
  raises `IndexError` on the supplied private save when a structure coordinate
  falls outside the fixed 100x100 heatmap grid.
- `examples/structure_api/ex_07_extract_structures_per_owner.py`
- `examples/json_api/ex_01_export_all_items.py` is implemented as
  `examples/export_all_items`, which writes save info plus all implemented
  domain JSON exports to an explicit output directory with a manifest.
- `examples/equipment_api/history/*.py` are local multi-save history utilities.
  They require a timestamped sequence of `.ark` snapshots or `.ark.gz`
  snapshots plus an `ark_files.json` manifest. The supplied private backup only
  provides a single current save per map, so private oracle execution is not
  available yet.

## Currently Blocked Read-Only Oracle Paths

These examples are offline-compatible in principle, but the supplied private
save currently drives upstream Python into malformed embedded cryopod parsing
before a stable aggregate can be returned. They remain useful implementation
targets, but are not reliable oracle cases until a quieter upstream invocation,
a fixture without the malformed cryopod payloads, or broader legacy/modded
cryopod support is available.

- `examples/dino_api/ex_10_get_cryopod_location.py`
- `examples/dino_api/ex_11_get_dino_pedigrees.py`
- `examples/structure_api/ex_03_create_structure_heatmap.py`

## Mutation-Copy Candidates

These examples mutate save data and must run only against copied inputs. They
are useful for write/reparse structural tests but cannot prove live Ark server
correctness.

- `examples/basic_parsing/ex_04_remove_blueprint_from_save_file.py`
- `examples/base_api/ex_02a_import_base_at_location.py`
- `examples/base_api/ex_02b_import_and_customize_base.py`
- `examples/dino_api/ex_06_change_dino_traits.py`
- `examples/dino_api/ex_12_remove_tamed_dino_from_save.py`
- `examples/dino_api/ex_13_extract_and_reinsert_dino.py`
- `examples/dino_api/ex_14_boost_dino_stats.py`
- `examples/dino_api/ex_15_force_grow_up_babies.py`
- `examples/equipment_api/ex_07_generate_blueprint_and_insert_in_save.py`
- `examples/structure_api/ex_05_modify_structures.py`
- `examples/structure_api/ex_08_modify_structures_of_tribe.py`

## Offline Skips

Network-facing behavior is intentionally unsupported in this Go port. Examples
that only use FTP to acquire a save remain offline oracle candidates after
patching the path to a local `.ark`; they are classified separately from true
network-only behavior.

Always skipped:

- `examples/rcon_api/ex_00_mirror_server_log.py`: live RCON subscription and
  polling.
- `examples/ex_ftp_config`: FTP config sample, not an offline parser example.

Skipped behavior inside otherwise useful examples:

- `ArkFtpClient.from_config(...).download_save_file(...)`: FTP acquisition
  boilerplate. For offline oracle use, replace the downloaded path with a local
  `.ark` fixture path before running the read-only/export/mutation-copy logic.
- `upload_save_file(...)`: live FTP upload is out of scope. Mutation examples
  may still be used only as copy/write/reparse structural tests against local
  throwaway files.
- `PlayerApi(..., ftp_config=...)` or constructor overloads that read FTP
  configuration: replace with local save/profile/tribe inputs where the example
  logic is otherwise offline-compatible.

## Local Cluster Candidates

Local cluster support is in scope only for local files. Upstream includes public
sample cluster files under `examples/player_api/cluster_data`, and examples
`ex_03_get_cluster_data.py` and `ex_04_get_cluster_data_directly.py` are useful
references for local-file parsing.
