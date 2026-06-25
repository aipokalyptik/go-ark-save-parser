# Oracle Comparison Summary

This file is safe to commit. It records only aggregate status for the
implemented offline Go example comparisons. Detailed oracle values, paths,
class names, stdout, and stderr stay in `.oracle/output/oracle-comparison.json`.

## Case Status

- `map_summary`: `pass` (summary metrics compared)
- `object_classes`: `pass` (class list compared)
- `object_summary`: `pass` (object-by-UUID byte and property counts compared)
- `property_positions`: `pass` (property metadata offsets and encoded byte counts compared)
- `class_lookup`: `pass` (storage class substring structure counts compared)
- `class_property_summary`: `pass` (class property-name aggregate compared)
- `export_json`: `pass` (save-info JSON metrics and class list compared)
- `local_profiles`: `pass` (local profile and tribe aggregate counts compared)
- `local_profile_player_aggregates`: `pass` (local player death and unlocked engram aggregates compared)
- `player_unlocked_engrams`: `pass` (save path unlocked-engram set compared)
- `player_list`: `pass` (save path player list aggregates compared)
- `tribe_list`: `pass` (save path tribe list aggregates compared)
- `player_all`: `pass` (save path player aggregate fallback compared)
- `player_tribe_links`: `pass` (player tribe active and inactive relation aggregates compared)
- `player_and_tribe_data`: `pass` (combined player, tribe, and relation JSON aggregates compared)
- `player_inventory`: `pass` (player inventory item count and location presence compared)
- `player_inventories`: `pass` (all-player inventory aggregate counts compared)
- `dino_filter`: `pass` (dino aggregate counts compared)
- `dino_best_stat_no_cryos`: `pass` (best stat dino without cryopods compared)
- `dino_best_base_stat`: `pass` (class-filtered tamed base stat dino compared)
- `dino_most_mutated`: `pass` (most mutated tamed dino aggregate compared)
- `dino_babies`: `pass` (wild and tamed baby dino counts compared)
- `dino_wild_tamables`: `pass` (wild and tameable dino counts compared)
- `dino_wild_tamed`: `pass` (wild-tamed dino count and max level compared)
- `dino_heatmap`: `pass` (direct dino heatmap cell aggregates compared without cryopod parsing)
- `property_filter`: `pass` (property-name filtered object counts compared)
- `stackable_count`: `pass` (stackable item count and total quantity compared)
- `stackable_owned_by`: `pass` (owned advanced rifle bullet count and total compared)
- `domain_json_stackables`: `pass` (stackable domain JSON aggregate counts compared)
- `equipment_longneck_blueprint_damage`: `pass` (longneck blueprint count and max damage compared)
- `equipment_best`: `pass` (highest weapon damage and armor durability values compared)
- `equipment_summary`: `pass` (canonical direct equipment class counts compared)
- `equipment_rank`: `pass` (stable high-rating non-crafted equipment rank aggregates compared; count and average-stat parity remain open)
- `equipment_ascendant_weapon_bps`: `pass` (ascendant weapon blueprint count and max damage compared)
- `equipment_saddles`: `pass` (direct saddle count compared; upstream cryopod saddle extraction blocked by malformed private cryopods and armor-value parity needs default armor tables)
- `equipment_owned_by`: `pass` (owned advanced weapon blueprint count and max damage compared)
- `structure_owner_count`: `pass` (owned structure count compared)
- `structure_owners`: `pass` (stable structure owner identity aggregates compared; selected row field counts can include extra inventory-bearing rows)
- `structure_owner_locations`: `pass` (structure owner/location multi-structure cells compared; exact owner and cell counts remain open under full structure parse performance)
- `structure_at_location`: `pass` (map-coordinate structure and connected counts compared)
- `base_components`: `pass` (connected base component aggregate counts compared)
- `domain_json_dinos`: `pass` (dino domain JSON aggregate counts compared)
- `cluster_json`: `pass` (local cluster upload counts compared)
- `local_tribute`: `pass` (local tribute aggregate counts compared)
- `tribute_json`: `pass` (local tribute JSON aggregate counts compared)
- `logging_config`: `pass` (standalone logging configuration output compared)

## Counts

- `pass`: 46
