// Package arkbinary contains low-level binary readers, name-table context, and
// decompression helpers for Ark save payloads.
//
// It handles little-endian primitives, Ark strings, UUIDs, indexed name lookup,
// zlib inflation, wildcard decompression, and embedded compressed archive
// decoding. Higher-level packages build on these primitives for save, archive,
// and property parsing.
package arkbinary
