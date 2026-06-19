# Examples

These examples are offline-only Go counterparts for runnable upstream example
categories. They require local files and do not use FTP, RCON, or live server
access.

Run an example with `go run`:

```sh
go run ./examples/map_summary /path/to/Valguero_WP.ark
```

## Read-Only Save Examples

- `map_summary`: open a local `.ark` and print map, version, object, and name
  counts. This covers the basic parsing and save-info JSON workflow.
- `object_classes`: list unique object class names from a local `.ark`, matching
  the class lookup/filtering style of the upstream basic parsing examples.

## Profile, Tribe, And Cluster Examples

- `local_profiles`: scan a directory for local `.arkprofile`, `.arktribe`, and
  extensionless local cluster files, then print discovered and parsed counts.
- `cluster_json`: read one local cluster file and print the cluster upload
  summary JSON.

## Mutation-Copy Example

- `mutation_copy`: copy a `.ark` save to a new explicit output path. Mutation
  helpers never modify inputs in place and are structurally tested only; live
  Ark server behavior remains unverified.
