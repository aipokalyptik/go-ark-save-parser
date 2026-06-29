// Package arkarchive parses local Unreal archive-style payloads used by player,
// tribe, local cluster, local tribute, and embedded modern cryopod data.
//
// Modern archive formats are parsed into object records with property
// containers and per-object property errors. Legacy archive object parsing is
// intentionally reported as unsupported until a concrete offline fixture and
// oracle case prove the exact behavior needed by the Go port.
package arkarchive
