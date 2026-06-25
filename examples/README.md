# Examples

These examples are offline-only Go counterparts for implemented runnable
upstream example categories. They require local files and do not use FTP, RCON,
or live server access.

Example output is not redacted. It can include local paths, class names,
profile/tribe IDs, upload IDs, locations, and other save-derived details. Treat
captured example output as private unless you have reviewed and sanitized it.
Use the `arksave --redact ...` CLI commands when you need safer aggregate output
for logs, issues, or reports.

Run an example with `go run`:

```sh
go run ./examples/map_summary /path/to/Valguero_WP.ark
```

## Read-Only Save Examples

- `map_summary`: open a local `.ark` and print map, version, object, and name
  counts. This covers the basic parsing and save-info JSON workflow.
- `parse_all`: open a local `.ark`, fault-tolerantly parse all game objects,
  and print object, parsed-object, and parse-fault counts.
- `object_classes`: list unique object class names from a local `.ark`, matching
  the class lookup/filtering style of the upstream basic parsing examples.
- `object_summary`: look up one object by UUID and print raw byte and parsed
  property counts without printing class names or property values.
- `property_positions`: look up one object by UUID and print aggregate property
  metadata counts for name offsets, value offsets, encoded byte spans, and
  nonzero property positions without printing property names or values.
- `class_lookup`: count objects and distinct classes matching one or more class
  name substrings without printing class names or object UUIDs.
- `class_property_summary`: parse objects matching one class substring and
  print aggregate object, unique-property, and parse-fault counts.
- `property_filter`: count objects and classes whose raw save object payloads
  contain any requested property name, matching the upstream property-name
  prefilter workflow without deep-parsing every object.
- `dino_filter`: parse local dino objects, run basic tamed/wild/cryopodded
  filters, and print aggregate class counts. Pass `--no-cryos` before the save
  path when comparing against no-cryopod oracle data.
- `dino_best_stat`: find the dino with the highest parsed stat points, or print
  `no_match` when no stat-bearing dino status components are present. Pass
  `--no-cryos` before the save path to skip cryopod payloads.
- `dino_best_base_stat`: find the tamed, direct-save dino with the highest base
  points for a requested blueprint and stat without printing object IDs,
  locations, or owners.
- `dino_most_mutated`: find the tamed dino with the highest upstream-compatible
  displayed mutation count and print only aggregate-safe values.
- `dino_babies`: count wild and tamed baby dinos without printing individual
  names, locations, or owners.
- `dino_wild_tamables`: count wild dinos and upstream-compatible wild-tamable
  dinos without printing class names or locations.
- `dino_wild_tamed`: count tamed dinos with no parsed ancestors and report the
  highest current level without printing names, classes, or owners.
- `dino_export_from_save`: write direct-save dino rows plus linked status and
  inventory rows, when present, to an explicit output directory and print
  dino/row/fault counts. Cryopod insertion, stat mutation, and live-server
  validation remain mutation-copy-only and unverified.
- `dino_heatmap`: generate a compact JSON summary for a local dino heatmap
  using an explicit output path, optional resolution, and optional `--no-cryos`
  mode for oracle comparisons that avoid malformed embedded cryopod payloads.
- `stackable_count`: filter resource, consumable, or ammo stackables by one or
  more explicit blueprint paths and print aggregate item and quantity counts.
- `stackable_owned_by`: filter stackables by blueprint and owning tribe through
  the structure inventory container relationship.
- `equipment_summary`: parse local equipment items and print aggregate counts
  for weapons, armor, direct saddles, modern cryopod saddles, and shields.
- `equipment_best`: mirror upstream read-only equipment examples by printing
  highest weapon damage and highest armor durability values over upstream
  canonical weapon and armor class lists.
- `equipment_ascendant_weapon_bps`: count ascendant weapon blueprints and print
  the maximum parsed damage value.
- `equipment_saddles`: count upstream-listed direct saddle items plus tolerant
  modern cryopod saddle parses and print aggregate saddle counts and max armor.
- `equipment_owned_by`: count weapon blueprints of one class owned by a target
  tribe through the structure inventory container relationship.
- `equipment_history`: read a JSON manifest of local `.ark` snapshots, compare
  equipment snapshots by stable item content, and write a JSON change report.
- `equipment_export_from_save`: write parsed equipment item rows to an
  explicit output directory and print item/row/fault counts. Inventory
  insertion, generated blueprint construction, and live-server validation
  remain mutation-copy-only and unverified.
- `structure_owner_count`: count local structure objects owned by a specific
  tribe ID.
- `structure_owners`: summarize parsed structure owner fields without printing
  individual owner names or IDs.
- `structure_owner_locations`: write an upstream-style owner/location JSON
  grouping for owned structures with map coordinates. Output can include
  save-derived owner labels and coordinates.
- `structure_at_location`: count structures near map coordinates and expand
  the result with directly connected structures.
- `structure_heatmap`: generate a compact JSON summary for a local structure
  heatmap using an explicit output path, resolution, and minimum-per-cell
  threshold.
- `base_components`: group parsed structures into linked connected components
  and print aggregate base counts using a selected-property scan over structure
  IDs and linked-structure UUIDs.
- `base_export_from_save`: write base metadata, raw structure rows, and
  structure location JSON files to an explicit output directory and print
  base/structure/fault counts. Inventory item expansion, binary import, and
  live-server validation remain mutation-copy-only and unverified.

## Profile, Tribe, And Cluster Examples

- `local_profiles`: scan a directory for local `.arkprofile`, `.arktribe`, and
  extensionless local cluster files, then print discovered counts, parsed
  counts, aggregate tribe-player links, deaths, levels, experience, and
  unlocked engram blueprint counts.
- `player_all`: accept a save path or save directory and print player/tribe
  aggregate counts, falling back to sibling local profile files when a save has
  no embedded player store.
- `player_list`: accept a save path or save directory and print privacy-safe
  player list aggregates for the upstream all-players iteration workflow.
- `tribe_list`: accept a save path or save directory and print privacy-safe
  tribe list aggregates for the upstream all-tribes iteration workflow.
- `player_unlocked_engrams`: accept a save path or save directory and print the
  sorted distinct unlocked engram blueprint set plus aggregate boundary values.
- `player_tribe_links`: accept a save path or save directory and print active
  tribe-player links plus inactive-member and missing-tribe aggregate counts.
- `player_and_tribe_data`: accept a save path or save directory and print a
  deterministic JSON summary for players, tribes, and active/inactive relation
  rows without private names or IDs.
- `player_inventory`: open a local `.ark`, resolve a player data ID through its
  pawn, and print the linked inventory item count plus pawn location when
  present.
- `player_inventories`: open a local `.ark`, resolve inventories for all
  save-contained players or sibling local profiles, and print privacy-safe
  aggregate inventory counts.
- `cluster_json`: read one local cluster file and print the cluster upload
  summary JSON.
- `local_tribute`: read local compact tribute index files and print aggregate
  player-data and tribe-data ID counts.
- `tribute_json`: read one local compact tribute index file or directory and
  print the tribute summary JSON.
- `export_all_items`: write save info plus all implemented domain JSON exports
  into an explicit output directory with a manifest.

## Utility Examples

- `logging_config`: demonstrate the Go logging helper's per-level filtering and
  all-level switch without reading save data or writing persistent config.

## Mutation-Copy Example

- `mutation_copy`: copy a `.ark` save to a new explicit output path, remove an
  object row or all objects whose class contains a substring from a copied save,
  import raw structure rows from `base_export_from_save`, import direct-save
  dino/status/inventory rows from `dino_export_from_save`, import equipment
  item rows from `equipment_export_from_save`, upsert object bytes from hex, or
  upsert a custom-table value from hex. Mutation helpers never modify inputs in
  place and are structurally tested only; live Ark server behavior remains
  unverified.
