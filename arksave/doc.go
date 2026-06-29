// Package arksave opens local SQLite-backed `.ark` map saves.
//
// It reads save headers, name tables, custom values, actor transforms, object
// class metadata, raw object bytes, parsed objects, selected-property scans,
// fault-tolerant object enumeration, and optional object-byte caching. The
// package is pure Go apart from the embedded modernc SQLite driver and does not
// require Python or system SQLite libraries.
package arksave
