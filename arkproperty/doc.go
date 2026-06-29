// Package arkproperty parses Unreal property records used inside Ark save
// objects and local archive payloads.
//
// The parser supports the property encodings exercised by the offline Go port,
// preserves encoded byte spans for structural mutation workflows, and keeps raw
// fallback data for unknown property or struct forms where possible.
package arkproperty
