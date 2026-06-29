// Package arktribute reads local tribute index files.
//
// It supports `.arktributetribe` and `.arktributetribetribe` files, returning
// player and tribe data IDs for offline inspection. Directory helpers can
// collect parse faults separately so batch tools can report malformed files
// without discarding valid tribute indexes.
package arktribute
