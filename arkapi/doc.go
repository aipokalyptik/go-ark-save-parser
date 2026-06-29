// Package arkapi provides high-level offline read APIs for local Ark Survival
// Ascended save data.
//
// The package wraps lower-level save, profile, tribe, cluster, tribute, and
// object parsers with typed summaries for common command-line workflows:
// players, tribes, dinos, structures, equipment, stackables, bases, JSON
// exports, heatmaps, and binary export helpers. APIs operate on local files and
// return explicit errors plus fault collections where best-effort parsing can
// preserve valid rows while reporting malformed objects.
package arkapi
