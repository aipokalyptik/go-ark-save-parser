# Examples

These examples are offline-only Go counterparts for runnable upstream example
categories. They require local files and do not use FTP, RCON, or live server
access.

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
- `object_classes`: list unique object class names from a local `.ark`, matching
  the class lookup/filtering style of the upstream basic parsing examples.
- `property_filter`: count objects and classes whose raw save object payloads
  contain any requested property name, matching the upstream property-name
  prefilter workflow without deep-parsing every object.
- `dino_filter`: parse local dino objects, run basic tamed/wild filters, and
  print aggregate class counts.
- `stackable_count`: filter resource, consumable, or ammo stackables by one or
  more explicit blueprint paths and print aggregate item and quantity counts.

## Profile, Tribe, And Cluster Examples

- `local_profiles`: scan a directory for local `.arkprofile`, `.arktribe`, and
  extensionless local cluster files, then print discovered counts, parsed
  counts, aggregate tribe-player links, deaths, levels, experience, and
  unlocked engram blueprint counts.
- `player_inventory`: open a local `.ark`, resolve a player data ID through its
  pawn, and print the linked inventory item count.
- `cluster_json`: read one local cluster file and print the cluster upload
  summary JSON.
- `local_tribute`: read local compact tribute index files and print aggregate
  player-data and tribe-data ID counts.

## Mutation-Copy Example

- `mutation_copy`: copy a `.ark` save to a new explicit output path. Mutation
  helpers never modify inputs in place and are structurally tested only; live
  Ark server behavior remains unverified.
