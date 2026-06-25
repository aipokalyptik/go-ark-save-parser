# Oracle Comparison Summary

This file is safe to commit. It records only aggregate status for the
implemented offline Go example comparisons. Detailed oracle values, paths,
class names, stdout, and stderr stay in `.oracle/output/oracle-comparison.json`.

## Case Status

- `map_summary`: `pass` (summary metrics compared)
- `object_classes`: `pass` (class list compared)
- `object_summary`: `pass` (object-by-UUID byte and property counts compared)
- `class_lookup`: `pass` (storage class substring structure counts compared)
- `class_property_summary`: `pass` (class property-name aggregate compared)
- `export_json`: `pass` (save-info JSON metrics and class list compared)
- `local_profiles`: `pass` (local profile and tribe aggregate counts compared)
- `local_profile_player_aggregates`: `pass` (local player death and unlocked engram aggregates compared)
- `player_all`: `pass` (save path player aggregate fallback compared)
- `player_tribe_links`: `pass` (player tribe active and inactive relation aggregates compared)
- `player_inventory`: `pass` (player inventory item count and location presence compared)
- `dino_filter`: `pass` (dino aggregate counts compared)
- `dino_best_stat_no_cryos`: `pass` (best stat dino without cryopods compared)
- `dino_best_base_stat`: `pass` (class-filtered tamed base stat dino compared)
- `dino_most_mutated`: `pass` (most mutated tamed dino aggregate compared)
- `dino_babies`: `pass` (wild and tamed baby dino counts compared)
- `dino_wild_tamables`: `pass` (wild and tameable dino counts compared)
- `dino_wild_tamed`: `pass` (wild-tamed dino count and max level compared)
- `property_filter`: `pass` (property-name filtered object counts compared)
- `stackable_count`: `pass` (stackable item count and total quantity compared)
- `stackable_owned_by`: `pass` (owned advanced rifle bullet count and total compared)
- `equipment_longneck_blueprint_damage`: `pass` (longneck blueprint count and max damage compared)
- `equipment_best`: `pass` (highest weapon damage and armor durability values compared)
- `equipment_ascendant_weapon_bps`: `pass` (ascendant weapon blueprint count and max damage compared)
- `equipment_saddles`: `pass` (direct saddle count compared; upstream cryopod saddle extraction blocked by malformed private cryopods and armor-value parity needs default armor tables)
- `equipment_owned_by`: `pass` (owned advanced weapon blueprint count and max damage compared)
- `structure_owner_count`: `pass` (owned structure count compared)
- `base_components`: `pass` (connected base component aggregate counts compared)
- `domain_json_dinos`: `pass` (dino domain JSON aggregate counts compared)
- `cluster_json`: `pass` (local cluster upload counts compared)
- `local_tribute`: `pass` (local tribute aggregate counts compared)

## Counts

- `pass`: 31
