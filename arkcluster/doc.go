// Package arkcluster reads local cluster upload files from disk.
//
// Cluster support is intentionally local-file-only. The package discovers
// extensionless cluster files, parses supported archive payloads, exposes typed
// uploaded item and dino summaries, and offers fault-preserving directory reads
// for tooling that should continue past malformed local uploads.
package arkcluster
