package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkmutation"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

const redactedValue = "[redacted]"

var (
	version = "dev"
	commit  = "unknown"
	builtAt = "unknown"
)

var ignoredEquipmentNameParts = []string{
	"WeaponCrossbow",
	"WeaponMetalHatchet",
	"WeaponMetalPick",
	"WeaponBow",
	"Chitin",
	"Hide",
	"WeaponPike",
	"WeaponGun",
	// Preserve the upstream example token exactly; the mixed-case typo is not a Cloth armor filter.
	"CLoth",
}

type runOptions struct {
	Redact   bool
	NoCryos  bool
	Help     bool
	Metadata bool
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	opts, args, err := splitOptions(args)
	if err != nil {
		return err
	}
	if opts.Help {
		return usage(out)
	}
	if opts.Metadata {
		return printVersion(out)
	}
	if len(args) == 0 {
		return usage(out)
	}
	switch args[0] {
	case "help":
		return usage(out)
	case "version":
		return printVersion(out)
	case "inspect":
		if len(args) != 2 {
			return fmt.Errorf("%s requires a local .ark path", args[0])
		}
		return inspect(args[1], out)
	case "parse":
		if len(args) != 2 {
			return fmt.Errorf("parse requires a local .ark path")
		}
		return parseSave(args[1], out, opts)
	case "map-summary":
		if len(args) != 2 {
			return fmt.Errorf("map-summary requires a local .ark path")
		}
		return mapSummary(args[1], out)
	case "object-classes":
		if len(args) != 2 {
			return fmt.Errorf("object-classes requires a local .ark path")
		}
		return objectClasses(args[1], out)
	case "object-summary":
		if len(args) != 3 {
			return fmt.Errorf("object-summary requires a local .ark path and object uuid")
		}
		return objectSummary(args[1], args[2], out)
	case "property-positions":
		if len(args) != 3 {
			return fmt.Errorf("property-positions requires a local .ark path and object uuid")
		}
		return propertyPositions(args[1], args[2], out)
	case "class-lookup":
		if len(args) < 3 {
			return fmt.Errorf("class-lookup requires a local .ark path and at least one class substring")
		}
		return classLookup(args[1], args[2:], out)
	case "class-property-summary":
		if len(args) != 3 {
			return fmt.Errorf("class-property-summary requires a local .ark path and class substring")
		}
		return classPropertySummary(args[1], args[2], out)
	case "property-filter":
		if len(args) < 3 {
			return fmt.Errorf("property-filter requires a local .ark path and at least one property name")
		}
		return propertyFilter(args[1], args[2:], out)
	case "structure-health":
		if len(args) != 2 {
			return fmt.Errorf("structure-health requires a local .ark path")
		}
		return structureHealth(args[1], out)
	case "structure-owner-count":
		if len(args) != 3 {
			return fmt.Errorf("structure-owner-count requires a local .ark path and tribe id")
		}
		return structureOwnerCount(args[1], args[2], out, opts)
	case "structure-owners":
		if len(args) != 2 {
			return fmt.Errorf("structure-owners requires a local .ark path")
		}
		return structureOwners(args[1], out)
	case "structure-owner-locations":
		if len(args) != 2 && len(args) != 3 && len(args) != 4 {
			return fmt.Errorf("structure-owner-locations requires a local .ark path with optional map and digits")
		}
		mapName := ""
		if len(args) >= 3 {
			mapName = args[2]
		}
		digits := 1
		if len(args) == 4 {
			value, err := strconv.Atoi(args[3])
			if err != nil {
				return fmt.Errorf("parse digits: %w", err)
			}
			digits = value
		}
		return structureOwnerLocations(args[1], mapName, digits, out, opts)
	case "structure-demolishable":
		return structureDemolishable(args[1:], out, opts)
	case "structure-heatmap":
		if len(args) < 3 || len(args) > 5 {
			return fmt.Errorf("structure-heatmap requires a local .ark path, explicit output path, optional resolution, and optional min-in-cell")
		}
		resolution := 100
		if len(args) >= 4 {
			value, err := strconv.Atoi(args[3])
			if err != nil {
				return fmt.Errorf("parse resolution: %w", err)
			}
			resolution = value
		}
		minInCell := 1
		if len(args) == 5 {
			value, err := strconv.Atoi(args[4])
			if err != nil {
				return fmt.Errorf("parse min-in-cell: %w", err)
			}
			minInCell = value
		}
		return structureHeatmap(args[1], args[2], resolution, minInCell, out)
	case "base-components":
		if len(args) != 2 {
			return fmt.Errorf("base-components requires a local .ark path")
		}
		return baseComponents(args[1], out)
	case "dinos":
		if len(args) != 2 {
			return fmt.Errorf("dinos requires a local .ark path")
		}
		return dinos(args[1], out)
	case "dino-wild-tamables":
		if len(args) != 2 {
			return fmt.Errorf("dino-wild-tamables requires a local .ark path")
		}
		return dinoWildTamables(args[1], out)
	case "dino-babies":
		if len(args) != 2 {
			return fmt.Errorf("dino-babies requires a local .ark path")
		}
		return dinoBabies(args[1], out)
	case "dino-best-stat":
		if len(args) != 2 {
			return fmt.Errorf("dino-best-stat requires a local .ark path")
		}
		return dinoBestStat(args[1], out)
	case "dino-best-base-stat":
		if len(args) != 4 {
			return fmt.Errorf("dino-best-base-stat requires a local .ark path, dino blueprint, and stat")
		}
		return dinoBestBaseStat(args[1], args[2], args[3], out)
	case "dino-most-mutated":
		if len(args) != 2 {
			return fmt.Errorf("dino-most-mutated requires a local .ark path")
		}
		return dinoMostMutated(args[1], out)
	case "dino-wild-tamed":
		if len(args) != 2 {
			return fmt.Errorf("dino-wild-tamed requires a local .ark path")
		}
		return dinoWildTamed(args[1], out)
	case "dino-claimable":
		return dinoClaimable(args[1:], out, opts)
	case "dino-heatmap":
		if len(args) < 3 || len(args) > 4 {
			return fmt.Errorf("dino-heatmap requires a local .ark path, explicit output path, and optional resolution")
		}
		resolution := 100
		if len(args) == 4 {
			value, err := strconv.Atoi(args[3])
			if err != nil {
				return fmt.Errorf("parse resolution: %w", err)
			}
			resolution = value
		}
		return dinoHeatmap(args[1], args[2], resolution, out, opts)
	case "equipment-summary":
		if len(args) != 2 {
			return fmt.Errorf("equipment-summary requires a local .ark path")
		}
		return equipmentSummary(args[1], out)
	case "equipment-saddles":
		if len(args) != 2 {
			return fmt.Errorf("equipment-saddles requires a local .ark path")
		}
		return equipmentSaddles(args[1], out)
	case "equipment-best":
		if len(args) != 2 {
			return fmt.Errorf("equipment-best requires a local .ark path")
		}
		return equipmentBest(args[1], out)
	case "equipment-rank":
		if len(args) != 2 {
			return fmt.Errorf("equipment-rank requires a local .ark path")
		}
		return equipmentRank(args[1], out)
	case "equipment-ascendant-weapon-bps":
		if len(args) != 2 {
			return fmt.Errorf("equipment-ascendant-weapon-bps requires a local .ark path")
		}
		return equipmentAscendantWeaponBPs(args[1], out)
	case "equipment-history":
		if len(args) != 3 {
			return fmt.Errorf("equipment-history requires a manifest JSON path and explicit output path")
		}
		return equipmentHistory(args[1], args[2], out)
	case "equipment-owned-by":
		if len(args) != 4 {
			return fmt.Errorf("equipment-owned-by requires a local .ark path, blueprint, and tribe id")
		}
		return equipmentOwnedBy(args[1], args[2], args[3], out, opts)
	case "stackables":
		if len(args) != 2 {
			return fmt.Errorf("stackables requires a local .ark path")
		}
		return stackables(args[1], out)
	case "stackable-owned-by":
		if len(args) != 4 {
			return fmt.Errorf("stackable-owned-by requires a local .ark path, blueprint, and tribe id")
		}
		return stackableOwnedBy(args[1], args[2], args[3], out, opts)
	case "player-inventories":
		if len(args) != 2 {
			return fmt.Errorf("player-inventories requires a local .ark path")
		}
		return playerInventories(args[1], out)
	case "player-roster":
		if len(args) != 2 {
			return fmt.Errorf("player-roster requires a local .ark path or save directory")
		}
		return playerRoster(args[1], out)
	case "tribe-roster":
		if len(args) != 2 {
			return fmt.Errorf("tribe-roster requires a local .ark path or save directory")
		}
		return tribeRoster(args[1], out)
	case "player-tribe-links":
		if len(args) != 2 {
			return fmt.Errorf("player-tribe-links requires a local .ark path or save directory")
		}
		return playerTribeLinks(args[1], out)
	case "players":
		if len(args) != 2 {
			return fmt.Errorf("players requires a local .arkprofile path")
		}
		return players(args[1], out, opts)
	case "tribes":
		if len(args) != 2 {
			return fmt.Errorf("tribes requires a local .arktribe path")
		}
		return tribes(args[1], out, opts)
	case "cluster":
		if len(args) != 2 {
			return fmt.Errorf("cluster requires a local cluster file or directory path")
		}
		return cluster(args[1], out, opts)
	case "cluster-summary":
		if len(args) != 2 {
			return fmt.Errorf("cluster-summary requires a local cluster file or directory path")
		}
		return clusterSummary(args[1], out, opts)
	case "tribute":
		if len(args) != 2 {
			return fmt.Errorf("tribute requires a local .arktributetribe file or directory path")
		}
		return tribute(args[1], out, opts)
	case "export-json":
		if len(args) != 3 {
			return fmt.Errorf("export-json requires a local .ark path and explicit output path")
		}
		return exportJSON(args[1], args[2], out, opts)
	case "export-domain-json":
		if len(args) != 4 {
			return fmt.Errorf("export-domain-json requires a local .ark path, domain, and explicit output path")
		}
		return exportDomainJSON(args[1], args[2], args[3], out, opts)
	case "export-cluster-json":
		if len(args) != 3 {
			return fmt.Errorf("export-cluster-json requires a local cluster file path and explicit output path")
		}
		return exportClusterJSON(args[1], args[2], out, opts)
	case "export-tribute-json":
		if len(args) != 3 {
			return fmt.Errorf("export-tribute-json requires a local tribute file or directory path and explicit output path")
		}
		return exportTributeJSON(args[1], args[2], out, opts)
	case "mutate":
		return mutate(args[1:], out, opts)
	case "ftp", "rcon":
		return fmt.Errorf("%s is unsupported: this is an offline-only parser", args[0])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func splitOptions(args []string) (runOptions, []string, error) {
	opts := runOptions{}
	filtered := make([]string, 0, len(args))
	seenCommand := false
	for _, arg := range args {
		switch arg {
		case "--help", "-h":
			opts.Help = true
		case "--version", "-V":
			opts.Metadata = true
		case "--redact":
			opts.Redact = true
		case "--no-cryos":
			opts.NoCryos = true
		default:
			if strings.HasPrefix(arg, "--") && !seenCommand {
				return opts, nil, fmt.Errorf("unknown option %q", arg)
			}
			seenCommand = true
			filtered = append(filtered, arg)
		}
	}
	return opts, filtered, nil
}

func usage(out io.Writer) error {
	_, err := fmt.Fprintln(out, `Usage:
  arksave help
  arksave --help
  arksave version
  arksave --version
  arksave [--redact] inspect <save.ark>
  arksave [--redact] parse <save.ark>
  arksave map-summary <save.ark>
  arksave object-classes <save.ark>
  arksave object-summary <save.ark> <object-uuid>
  arksave property-positions <save.ark> <object-uuid>
  arksave class-lookup <save.ark> <class-substring> [class-substring...]
  arksave class-property-summary <save.ark> <class-substring>
  arksave property-filter <save.ark> <property> [property...]
  arksave structure-health <save.ark>
  arksave [--redact] structure-owner-count <save.ark> <tribe-id>
  arksave structure-owners <save.ark>
  arksave [--redact] structure-owner-locations <save.ark> [map] [digits]
  arksave [--redact] structure-demolishable <save.ark> [--game-user-settings path] [--decay-multiplier n] [--decay-periods path] [--map name] [--json] [--group-bases]
  arksave structure-heatmap <save.ark> <out.json> [resolution] [min-in-cell]
  arksave base-components <save.ark>
  arksave dinos <save.ark>
  arksave dino-wild-tamables <save.ark>
  arksave dino-babies <save.ark>
  arksave dino-best-stat <save.ark>
  arksave dino-best-base-stat <save.ark> <dino-blueprint> <stat>
  arksave dino-most-mutated <save.ark>
  arksave dino-wild-tamed <save.ark>
  arksave [--redact] dino-claimable <save.ark> [--game-user-settings path] [--claim-multiplier n] [--claim-period seconds] [--map name] [--json] [--debug-fields]
  arksave [--no-cryos] dino-heatmap <save.ark> <out.json> [resolution]
  arksave equipment-summary <save.ark>
  arksave equipment-saddles <save.ark>
  arksave equipment-best <save.ark>
  arksave equipment-rank <save.ark>
  arksave equipment-ascendant-weapon-bps <save.ark>
  arksave equipment-history <ark-files.json> <out.json>
  arksave [--redact] equipment-owned-by <save.ark> <blueprint> <tribe-id>
  arksave stackables <save.ark>
  arksave [--redact] stackable-owned-by <save.ark> <blueprint> <tribe-id>
  arksave player-inventories <save.ark>
  arksave player-roster <save.ark-or-directory>
  arksave tribe-roster <save.ark-or-directory>
  arksave player-tribe-links <save.ark-or-directory>
  arksave [--redact] players <player.arkprofile-or-directory>
  arksave [--redact] tribes <tribe.arktribe-or-directory>
  arksave [--redact] cluster <cluster-file-or-directory>
  arksave [--redact] cluster-summary <cluster-file-or-directory>
  arksave [--redact] tribute <tribute-file-or-directory>
  arksave [--redact] export-json <save.ark> <out.json>
  arksave [--redact] export-domain-json <save.ark> <dinos|structures|equipment|stackables|players|tribes|bases> <out.json>
  arksave [--redact] export-cluster-json <cluster-file> <out.json>
  arksave [--redact] export-tribute-json <tribute-file-or-directory> <out.json>
  arksave [--redact] mutate copy <save.ark> <out.ark>
  arksave [--redact] mutate remove-object <save.ark> <out.ark> <uuid>
  arksave [--redact] mutate remove-class-contains <save.ark> <out.ark> <class-substring>
  arksave [--redact] mutate import-base-binary <save.ark> <out.ark> <base-export-dir>
  arksave [--redact] mutate import-structure-binary <save.ark> <out.ark> <structure-export-dir>
  arksave [--redact] mutate import-dino-binary <save.ark> <out.ark> <dino-export-dir>
  arksave [--redact] mutate import-equipment-binary <save.ark> <out.ark> <equipment-export-dir>
  arksave [--redact] mutate put-object-hex <save.ark> <out.ark> <uuid> <hex-value>
  arksave [--redact] mutate replace-object-property-hex <save.ark> <out.ark> <uuid> <property-name> <position> <hex-encoded-property>
  arksave [--redact] mutate put-custom <save.ark> <out.ark> <key> <hex-value>

Offline-only scope: FTP and RCON are intentionally unsupported.
replace-object-property-hex requires a full encoded property record, not only scalar payload bytes.
Use --redact to hide local paths and identifier/detail fields in command output and JSON exports.`)
	return err
}

func printVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "arksave version=%s commit=%s built_at=%s\n", version, commit, builtAt)
	return err
}

func inspect(path string, out io.Writer) error {
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	ids, err := save.ObjectIDs()
	if err != nil {
		return err
	}
	if save.Context == nil {
		return errors.New("save context is nil")
	}
	_, err = fmt.Fprintf(
		out,
		"Map: %s\nSave version: %d\nGame time: %.3f\nObjects: %d\n",
		save.Context.MapName,
		save.Context.SaveVersion,
		save.Context.GameTime,
		len(ids),
	)
	return err
}

func parseSave(path string, out io.Writer, opts runOptions) error {
	info, err := arkapi.ExportSaveInfoFromPath(path)
	if err != nil {
		return err
	}
	summary, _, err := arkapi.GeneralParseSummaryFromPath(path)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Save: %s\nMap: %s\nSave version: %d\nObjects: %d\nParsed objects: %d\nParse faults: %d\n",
		displayString(path, opts),
		info.MapName,
		info.SaveVersion,
		summary.Objects,
		summary.Parsed,
		summary.Faults,
	)
	return err
}

func mapSummary(path string, out io.Writer) error {
	info, err := arkapi.ExportSaveInfoFromPath(path)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"map=%s save_version=%d objects=%d names=%d\n",
		info.MapName,
		info.SaveVersion,
		info.ObjectCount,
		info.NameCount,
	)
	return err
}

func objectClasses(path string, out io.Writer) error {
	classes, err := arkapi.GeneralClassesFromPath(path)
	if err != nil {
		return err
	}
	for _, className := range classes {
		if _, err := fmt.Fprintln(out, className); err != nil {
			return err
		}
	}
	return nil
}

func objectSummary(path string, objectIDArg string, out io.Writer) error {
	objectID, err := uuid.Parse(objectIDArg)
	if err != nil {
		return fmt.Errorf("parse object uuid: %w", err)
	}
	summary, err := arkapi.GeneralObjectSummaryFromPath(path, objectID)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Exists: %t\nBytes: %d\nProperties: %d\n",
		summary.Exists,
		summary.Bytes,
		summary.Properties,
	)
	return err
}

func propertyPositions(path string, objectIDArg string, out io.Writer) error {
	objectID, err := uuid.Parse(objectIDArg)
	if err != nil {
		return fmt.Errorf("parse object uuid: %w", err)
	}
	summary, err := arkapi.GeneralPropertyPositionSummaryFromPath(path, objectID)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Exists: %t\nProperties: %d\nName offsets: %d\nValue offsets: %d\nEncoded: %d\nPositioned: %d\nOffsets OK: %d\n",
		summary.Exists,
		summary.Properties,
		summary.NameOffsets,
		summary.ValueOffsets,
		summary.Encoded,
		summary.Positioned,
		summary.OffsetsOK,
	)
	return err
}

func classLookup(path string, classSubstrings []string, out io.Writer) error {
	summary, faults, err := arkapi.GeneralClassLookupSummaryFromPath(path, classSubstrings)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Objects: %d\nClasses: %d\nParse faults: %d\n",
		summary.Objects,
		summary.Classes,
		len(faults),
	)
	return err
}

func classPropertySummary(path string, classSubstring string, out io.Writer) error {
	summary, faults, err := arkapi.GeneralClassPropertySummaryFromPath(path, classSubstring)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Objects: %d\nProperties: %d\nParse faults: %d\n",
		summary.Objects,
		summary.Properties,
		len(faults),
	)
	return err
}

func propertyFilter(path string, propertyNames []string, out io.Writer) error {
	summary, err := arkapi.GeneralPropertyFilterSummaryFromPath(path, propertyNames)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Objects: %d\nClasses: %d\n",
		summary.Objects,
		summary.Classes,
	)
	return err
}

func structureHealth(path string, out io.Writer) error {
	summary, faults, err := arkapi.StructureHealthSummaryFromPath(path)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Structures: %d\nWith health: %d\nDamaged: %d\nFully repaired: %d\nWithout max health: %d\nAverage health: %.1f%%\nMinimum health: %.1f%%\nMaximum health: %.1f%%\nParse faults: %d\n",
		summary.Structures,
		summary.WithHealth,
		summary.Damaged,
		summary.FullyRepaired,
		summary.WithoutMaxHealth,
		summary.AverageHealthPercent,
		summary.MinimumHealthPercent,
		summary.MaximumHealthPercent,
		len(faults),
	)
	return err
}

func structureOwnerCount(path string, tribeIDArg string, out io.Writer, opts runOptions) error {
	tribeID64, err := strconv.ParseInt(tribeIDArg, 10, 32)
	if err != nil {
		return fmt.Errorf("parse tribe id: %w", err)
	}
	summary, faults, err := arkapi.StructureTribeOwnershipSummaryFromPath(path, int32(tribeID64))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Tribe ID: %v\nStructures: %d\nParse faults: %d\n",
		displayInt(summary.TribeID, opts),
		summary.Structures,
		len(faults),
	)
	return err
}

func structureOwners(path string, out io.Writer) error {
	summary, faults, err := arkapi.StructureOwnerSummaryFromPath(path)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Structures: %d\nWith tribe ID: %d\nWith player ID: %d\nWith tribe name: %d\nWith player name: %d\nWith original placer ID: %d\nUnique tribes: %d\nUnique players: %d\nUnique original placers: %d\nParse faults: %d\n",
		summary.Structures,
		summary.WithTribeID,
		summary.WithPlayerID,
		summary.WithTribeName,
		summary.WithPlayerName,
		summary.WithOriginalPlacerID,
		summary.UniqueTribes,
		summary.UniquePlayers,
		summary.UniqueOriginalPlacers,
		len(faults),
	)
	return err
}

func structureOwnerLocations(path string, mapName string, digits int, out io.Writer, opts runOptions) error {
	export, _, err := arkapi.StructureOwnerLocationsFromPathWithFaults(path, mapName, digits)
	if err != nil {
		return err
	}
	printable := export
	if opts.Redact {
		printable.OwnersByLocation = make([]arkapi.StructureOwnerLocationData, len(export.OwnersByLocation))
		for i, owner := range export.OwnersByLocation {
			printable.OwnersByLocation[i] = owner
			printable.OwnersByLocation[i].Owner = redactedValue
		}
	}
	if _, err := fmt.Fprintf(
		out,
		"Structures: %d\nOwners: %d\nCells: %d\nNamed cells: %d\nMulti-structure cells: %d\nSkipped without owner: %d\nSkipped without location: %d\nParse faults: %d\n",
		export.Structures,
		export.Owners,
		export.Cells,
		export.NamedCells,
		export.MultiStructureCells,
		export.SkippedWithoutOwner,
		export.SkippedWithoutLocation,
		export.FaultCount,
	); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(printable.OwnersByLocation, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, string(encoded))
	return err
}

func dinoClaimable(args []string, out io.Writer, runOpts runOptions) error {
	path, opts, jsonOut, debugFields, err := parseDinoClaimableArgs(args)
	if err != nil {
		return err
	}
	if debugFields {
		debug, err := arkapi.DinoClaimableFieldDebugFromPath(path)
		if err != nil {
			return err
		}
		if jsonOut {
			raw, err := json.MarshalIndent(debug, "", "  ")
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(out, string(raw))
			return err
		}
		return printDinoClaimableFieldDebug(debug, out)
	}
	report, _, err := arkapi.DinoClaimableReportFromPath(path, opts)
	if err != nil {
		return err
	}
	if runOpts.Redact {
		redactDinoClaimableReport(&report)
	}
	if jsonOut {
		raw, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, string(raw))
		return err
	}
	if _, err := fmt.Fprintf(
		out,
		"Computed offline dino claim eligibility using LastInAllyRangeSerialized with LastInAllyRangeTimeSerialized/TamedTimeStamp fallback\nDinos: %d\nOwned dinos: %d\nClaimable: %d\nUnknown timestamps: %d\nMultiplier: %.3f\nParse faults: %d\n",
		report.Summary.TotalDinos,
		report.Summary.OwnedDinos,
		report.Summary.ClaimableDinos,
		report.Summary.UnknownTimestampDinos,
		report.Summary.ClaimMultiplier,
		report.Summary.FaultCount,
	); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "OWNER\tLOCATION\tSPECIES\tNAME\tELAPSED\tREMAINING"); err != nil {
		return err
	}
	for _, row := range report.Dinos {
		if _, err := fmt.Fprintf(
			out,
			"%s\t%s\t%s\t%s\t%s\t%s\n",
			dinoClaimableOwnerDisplay(row.Owner),
			demolishableLocationDisplay(row.Location),
			row.ShortName,
			dinoNameDisplay(row.TamedName),
			formatSeconds(row.ElapsedSeconds),
			formatSeconds(row.RemainingSeconds),
		); err != nil {
			return err
		}
	}
	return nil
}

func printDinoClaimableFieldDebug(debug arkapi.DinoClaimableFieldDebug, out io.Writer) error {
	if _, err := fmt.Fprintf(out, "Dino candidates: %d\nParse faults: %d\n", debug.TotalDinoCandidates, debug.FaultCount); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "FIELD\tCOUNT"); err != nil {
		return err
	}
	for _, name := range debug.CandidateNames {
		if count := debug.FieldCounts[name]; count > 0 {
			if _, err := fmt.Fprintf(out, "%s\t%d\n", name, count); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseDinoClaimableArgs(args []string) (string, arkapi.DinoClaimableOptions, bool, bool, error) {
	if len(args) == 0 {
		return "", arkapi.DinoClaimableOptions{}, false, false, fmt.Errorf("dino-claimable requires a local .ark path")
	}
	path := args[0]
	opts := arkapi.DinoClaimableOptions{}
	jsonOut := false
	debugFields := false
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOut = true
		case "--debug-fields":
			debugFields = true
		case "--map":
			i++
			if i >= len(args) {
				return "", opts, false, false, fmt.Errorf("--map requires a value")
			}
			opts.MapName = args[i]
		case "--game-user-settings":
			i++
			if i >= len(args) {
				return "", opts, false, false, fmt.Errorf("--game-user-settings requires a path")
			}
			opts.GameUserSettingsPath = args[i]
		case "--claim-multiplier":
			i++
			if i >= len(args) {
				return "", opts, false, false, fmt.Errorf("--claim-multiplier requires a value")
			}
			value, err := strconv.ParseFloat(args[i], 64)
			if err != nil {
				return "", opts, false, false, fmt.Errorf("parse claim multiplier: %w", err)
			}
			if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
				return "", opts, false, false, fmt.Errorf("claim multiplier must be a positive finite number")
			}
			opts.ClaimMultiplier = value
		case "--claim-period":
			i++
			if i >= len(args) {
				return "", opts, false, false, fmt.Errorf("--claim-period requires a value")
			}
			value, err := strconv.ParseFloat(args[i], 64)
			if err != nil {
				return "", opts, false, false, fmt.Errorf("parse claim period: %w", err)
			}
			if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
				return "", opts, false, false, fmt.Errorf("claim period must be a positive finite number")
			}
			opts.ClaimPeriodSeconds = value
		default:
			if strings.HasPrefix(args[i], "--") {
				return "", opts, false, false, fmt.Errorf("unknown dino-claimable option %q", args[i])
			}
			return "", opts, false, false, fmt.Errorf("unexpected dino-claimable argument %q", args[i])
		}
	}
	return path, opts, jsonOut, debugFields, nil
}

func redactDinoClaimableReport(report *arkapi.DinoClaimableReport) {
	for i := range report.Dinos {
		report.Dinos[i].UUID = redactedValue
		report.Dinos[i].DinoID1 = 0
		report.Dinos[i].DinoID2 = 0
		if report.Dinos[i].TamedName != "" {
			report.Dinos[i].TamedName = redactedValue
		}
		report.Dinos[i].Owner = redactedDinoClaimableOwner(report.Dinos[i].Owner)
	}
}

func redactedDinoClaimableOwner(owner arkapi.DinoClaimableOwner) arkapi.DinoClaimableOwner {
	owner.SortKey = redactedValue
	if owner.TribeName != "" {
		owner.TribeName = redactedValue
	}
	if owner.TamerTribeID != 0 {
		owner.TamerTribeID = 0
	}
	if owner.TamerString != "" {
		owner.TamerString = redactedValue
	}
	if owner.PlayerName != "" {
		owner.PlayerName = redactedValue
	}
	if owner.PlayerID != 0 {
		owner.PlayerID = 0
	}
	if owner.TargetTeam != 0 {
		owner.TargetTeam = 0
	}
	return owner
}

func dinoClaimableOwnerDisplay(owner arkapi.DinoClaimableOwner) string {
	switch {
	case owner.TribeName != "":
		return owner.TribeName
	case owner.TamerString != "":
		return owner.TamerString
	case owner.TamerTribeID != 0:
		return strconv.FormatInt(int64(owner.TamerTribeID), 10)
	case owner.TargetTeam != 0:
		return strconv.FormatInt(int64(owner.TargetTeam), 10)
	case owner.PlayerName != "":
		return owner.PlayerName
	case owner.PlayerID != 0:
		return strconv.FormatInt(int64(owner.PlayerID), 10)
	default:
		return owner.SortKey
	}
}

func dinoNameDisplay(name string) string {
	if name == "" {
		return "-"
	}
	return name
}

func structureDemolishable(args []string, out io.Writer, runOpts runOptions) error {
	path, opts, jsonOut, err := parseStructureDemolishableArgs(args)
	if err != nil {
		return err
	}
	report, _, err := arkapi.StructureDemolishableReportFromPath(path, opts)
	if err != nil {
		return err
	}
	if runOpts.Redact {
		redactStructureDemolishableReport(&report)
	}
	if jsonOut {
		raw, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, string(raw))
		return err
	}
	if _, err := fmt.Fprintf(
		out,
		"Computed offline demolish eligibility using LastInAllyRangeTimeSerialized with LastEnterStasisTime fallback\nStructures: %d\nEligible: %d\nUnknown timestamps: %d\nMultiplier: %.3f\nParse faults: %d\n",
		report.Summary.TotalStructures,
		report.Summary.EligibleStructures,
		report.Summary.UnknownTimestampStructures,
		report.Summary.DecayMultiplier,
		report.Summary.FaultCount,
	); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "OWNER\tLOCATION\tSTRUCTURE\tTIER\tELAPSED\tREMAINING"); err != nil {
		return err
	}
	for _, row := range report.Structures {
		if _, err := fmt.Fprintf(
			out,
			"%s\t%s\t%s\t%s\t%s\t%s\n",
			demolishableOwnerDisplay(row.Owner),
			demolishableLocationDisplay(row.Location),
			row.ShortName,
			row.Tier,
			formatSeconds(row.ElapsedSeconds),
			formatSeconds(row.RemainingSeconds),
		); err != nil {
			return err
		}
	}
	if len(report.Bases) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(out, "\nBASE OWNER\tLOCATION\tELIGIBLE\tTOTAL\tOLDEST\tDOMINANT TIER"); err != nil {
		return err
	}
	for _, base := range report.Bases {
		if _, err := fmt.Fprintf(
			out,
			"%s\t%s\t%d\t%d\t%s\t%s\n",
			demolishableOwnerDisplay(base.Owner),
			demolishableLocationDisplay(base.AverageLocation),
			base.EligibleStructures,
			base.TotalStructures,
			formatSeconds(base.OldestElapsed),
			base.DominantTier,
		); err != nil {
			return err
		}
	}
	return nil
}

func parseStructureDemolishableArgs(args []string) (string, arkapi.StructureDemolishableOptions, bool, error) {
	if len(args) == 0 {
		return "", arkapi.StructureDemolishableOptions{}, false, fmt.Errorf("structure-demolishable requires a local .ark path")
	}
	path := args[0]
	opts := arkapi.StructureDemolishableOptions{}
	jsonOut := false
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOut = true
		case "--group-bases":
			opts.GroupBases = true
		case "--map":
			i++
			if i >= len(args) {
				return "", opts, false, fmt.Errorf("--map requires a value")
			}
			opts.MapName = args[i]
		case "--game-user-settings":
			i++
			if i >= len(args) {
				return "", opts, false, fmt.Errorf("--game-user-settings requires a path")
			}
			opts.GameUserSettingsPath = args[i]
		case "--decay-periods":
			i++
			if i >= len(args) {
				return "", opts, false, fmt.Errorf("--decay-periods requires a path")
			}
			opts.DecayPeriodsPath = args[i]
		case "--decay-multiplier":
			i++
			if i >= len(args) {
				return "", opts, false, fmt.Errorf("--decay-multiplier requires a value")
			}
			value, err := strconv.ParseFloat(args[i], 64)
			if err != nil {
				return "", opts, false, fmt.Errorf("parse decay multiplier: %w", err)
			}
			opts.DecayMultiplier = value
		default:
			if strings.HasPrefix(args[i], "--") {
				return "", opts, false, fmt.Errorf("unknown structure-demolishable option %q", args[i])
			}
			return "", opts, false, fmt.Errorf("unexpected structure-demolishable argument %q", args[i])
		}
	}
	return path, opts, jsonOut, nil
}

func redactStructureDemolishableReport(report *arkapi.StructureDemolishableReport) {
	for i := range report.Structures {
		report.Structures[i].UUID = redactedValue
		report.Structures[i].Owner = redactedStructureDemolishableOwner(report.Structures[i].Owner)
	}
	for i := range report.Bases {
		report.Bases[i].Owner = redactedStructureDemolishableOwner(report.Bases[i].Owner)
		for j := range report.Bases[i].StructureUUIDs {
			report.Bases[i].StructureUUIDs[j] = redactedValue
		}
	}
}

func redactedStructureDemolishableOwner(owner arkapi.StructureDemolishableOwner) arkapi.StructureDemolishableOwner {
	owner.SortKey = redactedValue
	if owner.TribeName != "" {
		owner.TribeName = redactedValue
	}
	if owner.TribeID != 0 {
		owner.TribeID = 0
	}
	if owner.PlayerName != "" {
		owner.PlayerName = redactedValue
	}
	if owner.PlayerID != 0 {
		owner.PlayerID = 0
	}
	if owner.OriginalPlacerID != 0 {
		owner.OriginalPlacerID = 0
	}
	return owner
}

func demolishableOwnerDisplay(owner arkapi.StructureDemolishableOwner) string {
	switch {
	case owner.TribeName != "":
		return owner.TribeName
	case owner.TribeID != 0:
		return strconv.FormatInt(int64(owner.TribeID), 10)
	case owner.PlayerName != "":
		return owner.PlayerName
	case owner.PlayerID != 0:
		return strconv.FormatInt(int64(owner.PlayerID), 10)
	default:
		return owner.SortKey
	}
}

func demolishableLocationDisplay(location *arkobject.MapCoords) string {
	if location == nil || location.InCryopod {
		return "unknown"
	}
	return fmt.Sprintf("%.2f,%.2f", location.Lat, location.Long)
}

func formatSeconds(seconds float64) string {
	if math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return "unknown"
	}
	return fmt.Sprintf("%.0fs", seconds)
}

func structureHeatmap(path string, outPath string, resolution int, minInCell int, out io.Writer) error {
	summary, err := arkapi.ExportStructureSelectedHeatmapSummaryJSONFromPath(path, outPath, arkapi.StructureHeatmapOptions{
		Resolution:   resolution,
		MinInSection: minInCell,
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Cells: %d\nTotal: %d\nMax: %d\nParse faults: %d\nSkipped coordinates: %d\nWrote: %s\n",
		summary.NonzeroCells,
		summary.Total,
		summary.Max,
		summary.Faults,
		summary.SkippedCoordinates,
		outPath,
	)
	return err
}

func baseComponents(path string, out io.Writer) error {
	stats, err := arkapi.BaseComponentStatsFromPath(path, "")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Components: %d\nTotal structures: %d\nLargest component: %d\nComponents at least 10: %d\nParse faults: %d\n",
		stats.Components,
		stats.TotalStructures,
		stats.LargestComponent,
		stats.ComponentsAtLeast10,
		stats.Faults,
	)
	return err
}

func dinos(path string, out io.Writer) error {
	summary, faults, err := arkapi.DinoPopulationSummaryFromPath(path, true)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Dinos: %d\nTamed: %d\nWild: %d\nCryopodded: %d\nClasses: %d\nParse faults: %d\n",
		summary.Dinos,
		summary.Tamed,
		summary.Wild,
		summary.Cryopodded,
		summary.Classes,
		len(faults),
	)
	return err
}

func dinoWildTamables(path string, out io.Writer) error {
	summary, faults, err := arkapi.DinoWildTamableSummaryFromPath(path)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Wild dinos: %d\nWild tamables: %d\nParse faults: %d\n",
		summary.WildDinos,
		summary.WildTamables,
		len(faults),
	)
	return err
}

func dinoBabies(path string, out io.Writer) error {
	summary, faults, err := arkapi.DinoBabySummaryFromPath(path, arkapi.BabyFilterOptions{
		IncludeTamed:      true,
		IncludeCryopodded: true,
		IncludeWild:       true,
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Baby dinos: %d\nTamed babies: %d\nWild babies: %d\nParse faults: %d\n",
		summary.Tamed+summary.Wild,
		summary.Tamed,
		summary.Wild,
		len(faults),
	)
	return err
}

func dinoBestStat(path string, out io.Writer) error {
	summary, faults, err := arkapi.DinoBestStatSummaryFromPath(path, arkapi.DinoBestStatOptions{})
	if err != nil {
		return err
	}
	if !summary.Found {
		_, err = fmt.Fprintf(out, "Best stat: none\nParse faults: %d\n", len(faults))
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Best stat: %s\nPoints: %d\nLevel: %d\nBlueprint: %s\nParse faults: %d\n",
		summary.Stat.String(),
		summary.Points,
		summary.Level,
		summary.Blueprint,
		len(faults),
	)
	return err
}

func dinoBestBaseStat(path string, blueprint string, statName string, out io.Writer) error {
	stat, ok := arkobject.DinoStatFromString(statName)
	if !ok {
		return fmt.Errorf("unknown dino stat %q", statName)
	}
	summary, faults, err := arkapi.DinoBestStatSummaryFromPath(path, arkapi.DinoBestStatOptions{
		Blueprints:      []string{blueprint},
		Stats:           []arkobject.DinoStat{stat},
		OnlyTamed:       true,
		ExcludeCryopods: true,
		BaseStat:        true,
	})
	if err != nil {
		return err
	}
	if !summary.Found {
		_, err = fmt.Fprintf(out, "Best base stat: none\nParse faults: %d\n", len(faults))
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Best base stat: %s\nPoints: %d\nLevel: %d\nBlueprint: %s\nParse faults: %d\n",
		summary.Stat.String(),
		summary.Points,
		summary.Level,
		summary.Blueprint,
		len(faults),
	)
	return err
}

func dinoMostMutated(path string, out io.Writer) error {
	summary, err := arkapi.DinoMostMutatedSummaryFromPath(path)
	if err != nil {
		return err
	}
	if !summary.Found {
		_, err = fmt.Fprintln(out, "Most mutated: none")
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Most mutated: %s\nTotal mutation points: %d\nMutation pairs: %d\nLevel: %d\n",
		summary.Blueprint,
		summary.TotalMutationPoints,
		summary.MutationPairs,
		summary.Level,
	)
	return err
}

func dinoWildTamed(path string, out io.Writer) error {
	summary, faults, err := arkapi.DinoWildTamedSummaryFromPath(path)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Wild-tamed dinos: %d\nMax level: %d\nParse faults: %d\n",
		summary.Dinos,
		summary.MaxLevel,
		len(faults),
	)
	return err
}

func dinoHeatmap(path string, outPath string, resolution int, out io.Writer, opts runOptions) error {
	summary, err := arkapi.ExportDinoHeatmapSummaryJSONFromPath(path, outPath, arkapi.DinoHeatmapOptions{
		Resolution:        resolution,
		IncludeCryopodded: !opts.NoCryos,
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Cells: %d\nTotal: %d\nMax: %d\nParse faults: %d\nSkipped coordinates: %d\nWrote: %s\n",
		summary.NonzeroCells,
		summary.Total,
		summary.Max,
		summary.Faults,
		summary.SkippedCoordinates,
		outPath,
	)
	return err
}

func equipmentSummary(path string, out io.Writer) error {
	summary, faults, err := arkapi.EquipmentSummaryFromPath(path, arkapi.EquipmentFilterOptions{})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Items: %d\nTotal quantity: %d\nAverage quantity: %.2f\nWeapon items: %d\nArmor items: %d\nSaddle items: %d\nShield items: %d\nBlueprints: %d\nEquipped: %d\nCrafted: %d\nWith custom data: %d\nCustom data entries: %d\nClasses: %d\nMax quality: %d\nTotal rating: %.2f\nAverage rating: %.2f\nMax rating: %.1f\nMax damage: %.1f\nMax armor: %.1f\nMax durability: %.1f\nParse faults: %d\n",
		summary.Items,
		summary.TotalQuantity,
		summary.AverageQuantity,
		summary.ByKind[arkobject.EquipmentWeapon],
		summary.ByKind[arkobject.EquipmentArmor],
		summary.ByKind[arkobject.EquipmentSaddle],
		summary.ByKind[arkobject.EquipmentShield],
		summary.Blueprints,
		summary.Equipped,
		summary.Crafted,
		summary.WithCustomData,
		summary.CustomDataEntries,
		summary.Classes,
		summary.MaxQuality,
		summary.TotalRating,
		summary.AverageRating,
		summary.MaxRating,
		summary.MaxDamage,
		summary.MaxArmor,
		summary.MaxCurrentDurability,
		len(faults),
	)
	return err
}

func equipmentSaddles(path string, out io.Writer) error {
	summary, faults, err := arkapi.EquipmentSaddleSummaryFromPath(path)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Item saddles: %d\nCryopod saddles: %d\nTotal saddles: %d\nMax armor: %.1f\nParse faults: %d\n",
		summary.ItemSaddles,
		summary.CryopodSaddles,
		summary.TotalSaddles,
		summary.MaxArmor,
		len(faults),
	)
	return err
}

func equipmentBest(path string, out io.Writer) error {
	summary, faults, err := arkapi.EquipmentBestSummaryFromPath(path)
	if err != nil {
		return err
	}
	if summary.WeaponFound {
		if _, err := fmt.Fprintf(
			out,
			"Best weapon damage: %.1f\nBest weapon: %s\nBest weapon crafted: %t\n",
			summary.Weapon.Stats.Damage,
			arkobject.ShortNameFromBlueprint(summary.Weapon.Blueprint),
			summary.Weapon.IsCrafted(),
		); err != nil {
			return err
		}
	} else if _, err := fmt.Fprintln(out, "Best weapon: none"); err != nil {
		return err
	}
	if summary.ArmorFound {
		if _, err := fmt.Fprintf(
			out,
			"Best armor durability: %.1f\nBest armor: %s\nBest armor crafted: %t\n",
			summary.Armor.Stats.Durability,
			arkobject.ShortNameFromBlueprint(summary.Armor.Blueprint),
			summary.Armor.IsCrafted(),
		); err != nil {
			return err
		}
	} else if _, err := fmt.Fprintln(out, "Best armor: none"); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Parse faults: %d\n", len(faults))
	return err
}

func equipmentRank(path string, out io.Writer) error {
	stats, faults, err := arkapi.EquipmentRankStatsFromPath(path, arkapi.EquipmentRankOptions{
		MinRating:        3,
		ExcludeCrafted:   true,
		IgnoredNameParts: ignoredEquipmentNameParts,
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Ranked: %d\nBest rating: %.1f\nBest average stat: %.1f\nCrafted: %d\nBlueprints: %d\nClasses: %d\nParse faults: %d\n",
		stats.Ranked,
		stats.BestRating,
		stats.BestAverageStat,
		stats.Crafted,
		stats.Blueprints,
		stats.Classes,
		len(faults),
	)
	return err
}

func equipmentAscendantWeaponBPs(path string, out io.Writer) error {
	summary, faults, err := arkapi.EquipmentSummaryFromPath(path, arkapi.EquipmentFilterOptions{
		Kinds:          []arkobject.EquipmentKind{arkobject.EquipmentWeapon},
		Blueprints:     arkapi.UpstreamWeaponBlueprints(),
		OnlyBlueprints: true,
		MinQuality:     arkapi.AscendantQualityIndex,
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Items: %d\nMax damage: %.1f\nParse faults: %d\n",
		summary.Items,
		summary.MaxDamage,
		len(faults),
	)
	return err
}

func equipmentHistory(manifestPath string, outputPath string, out io.Writer) error {
	report, err := arkapi.EquipmentHistoryReportFromManifestPath(manifestPath)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, append(data, '\n'), 0o600); err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"saves=%d initial=%d changes=%d final=%d wrote=%s\n",
		report.Saves,
		report.InitialCount,
		len(report.Changes),
		report.FinalCount,
		outputPath,
	)
	return err
}

func equipmentOwnedBy(path string, blueprint string, tribeIDArg string, out io.Writer, opts runOptions) error {
	tribeID64, err := strconv.ParseInt(tribeIDArg, 10, 32)
	if err != nil {
		return fmt.Errorf("parse tribe id: %w", err)
	}
	summary, faults, err := arkapi.EquipmentOwnedSummaryFromPath(path, arkapi.EquipmentFilterOptions{
		Kinds:          []arkobject.EquipmentKind{arkobject.EquipmentWeapon},
		Blueprints:     []string{blueprint},
		OnlyBlueprints: true,
	}, arkobject.ObjectOwner{TribeID: int32(tribeID64)})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Tribe ID: %v\nBlueprint: %s\nItems: %d\nMax damage: %.1f\nParse faults: %d\n",
		displayInt(int32(tribeID64), opts),
		displayString(blueprint, opts),
		summary.Items,
		summary.MaxDamage,
		len(faults),
	)
	return err
}

func stackables(path string, out io.Writer) error {
	summary, faults, err := arkapi.StackableSummaryFromPathWithFaults(path)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Stackable items: %d\nTotal quantity: %d\nParse faults: %d\n",
		summary.Items,
		summary.TotalQuantity,
		len(faults),
	)
	return err
}

func stackableOwnedBy(path string, blueprint string, tribeIDArg string, out io.Writer, opts runOptions) error {
	tribeID64, err := strconv.ParseInt(tribeIDArg, 10, 32)
	if err != nil {
		return fmt.Errorf("parse tribe id: %w", err)
	}
	summary, err := arkapi.StackableOwnedSummaryFromPath(path, []string{blueprint}, arkobject.ObjectOwner{TribeID: int32(tribeID64)})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Tribe ID: %v\nBlueprint: %s\nItems: %d\nTotal quantity: %d\n",
		displayInt(int32(tribeID64), opts),
		displayString(blueprint, opts),
		summary.Items,
		summary.TotalQuantity,
	)
	return err
}

func playerInventories(path string, out io.Writer) error {
	summary, faults, err := arkapi.PlayerInventorySummaryFromPath(path)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Players: %d\nWith inventory: %d\nWithout inventory: %d\nTotal items: %d\nMax items: %d\nMin items: %d\nAverage items: %.2f\nInventory faults: %d\n",
		summary.Players,
		summary.WithInventory,
		summary.WithoutInventory,
		summary.TotalItems,
		summary.MaxItems,
		summary.MinItems,
		summary.AverageItems,
		len(faults),
	)
	return err
}

func playerRoster(path string, out io.Writer) error {
	summary, faults, err := arkapi.PlayerRosterSummaryFromPath(path)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Players: %d\nWith names: %d\nHighest level: %d\nParse faults: %d\n",
		summary.Players,
		summary.WithNames,
		summary.HighestLevel,
		len(faults),
	)
	return err
}

func tribeRoster(path string, out io.Writer) error {
	summary, faults, err := arkapi.TribeRosterSummaryFromPath(path)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Tribes: %d\nWith names: %d\nMembers: %d\nDinos: %d\nParse faults: %d\n",
		summary.Tribes,
		summary.WithNames,
		summary.Members,
		summary.Dinos,
		len(faults),
	)
	return err
}

func playerTribeLinks(path string, out io.Writer) error {
	summary, err := arkapi.TribePlayerRelationSummaryFromPath(path)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Players: %d\nTribes: %d\nActive links: %d\nInactive members: %d\nPlayers without tribe: %d\nTribes with inactive: %d\nTribes without active: %d\n",
		summary.Players,
		summary.Tribes,
		summary.ActiveLinks,
		summary.InactiveMembers,
		summary.PlayersWithoutTribe,
		summary.TribesWithInactive,
		summary.TribesWithoutActive,
	)
	return err
}

func players(path string, out io.Writer, opts runOptions) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return playersDirectory(path, out, opts)
	}

	summary, err := arkapi.PlayerProfileFileSummaryFromPath(path)
	if err != nil {
		if summary.Archive.Path == "" {
			return err
		}
		if archiveErr := printArchiveSummary(out, "Player profile", summary.Archive, opts); archiveErr != nil {
			return archiveErr
		}
		return fmt.Errorf("parse player profile details: %w", err)
	}
	if err := printArchiveSummary(out, "Player profile", summary.Archive, opts); err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Character name: %s\nPlayer name: %s\nPlayer data ID: %v\nTribe ID: %v\nDeaths: %d\n",
		displayString(summary.Player.CharacterName, opts),
		displayString(summary.Player.PlayerName, opts),
		displayInt(summary.Player.PlayerDataID, opts),
		displayInt(summary.Player.TribeID, opts),
		summary.Player.NumDeaths,
	)
	return err
}

func playersDirectory(path string, out io.Writer, opts runOptions) error {
	summary, err := arkapi.PlayerDirectorySummaryFromPath(path)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(
		out,
		"Player directory: %s\nProfiles: %d\nPlayers: %d\nTotal deaths: %d\nAverage deaths: %.2f\nTotal level: %d\nAverage level: %.2f\nHighest level: %d\nTotal experience: %.2f\nAverage experience: %.2f\nHighest experience: %.2f\nTotal engram points: %d\nUnlocked engrams: %d\n",
		displayString(path, opts),
		summary.Files,
		len(summary.Players),
		summary.TotalDeaths,
		optionalFloat(summary.AverageDeaths, summary.HasAverageDeaths),
		summary.TotalLevel,
		optionalFloat(summary.AverageLevel, summary.HasAverageLevel),
		summary.HighestLevel,
		summary.TotalExperience,
		optionalFloat(summary.AverageExperience, summary.HasAverageExperience),
		summary.HighestExperience,
		summary.TotalEngramPoints,
		summary.UnlockedEngrams,
	); err != nil {
		return err
	}
	if opts.Redact {
		return nil
	}
	for _, player := range summary.Players {
		if _, err := fmt.Fprintf(
			out,
			"  player id=%d character=%s platform=%s tribe=%d level=%d deaths=%d\n",
			player.PlayerDataID,
			player.CharacterName,
			player.PlayerName,
			player.TribeID,
			player.Level,
			player.NumDeaths,
		); err != nil {
			return err
		}
	}
	return nil
}

func tribes(path string, out io.Writer, opts runOptions) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return tribesDirectory(path, out, opts)
	}

	fileSummary, err := arkapi.TribeFileSummaryFromPath(path)
	if err != nil {
		if fileSummary.Archive.Path == "" {
			return err
		}
		if archiveErr := printArchiveSummary(out, "Tribe save", fileSummary.Archive, opts); archiveErr != nil {
			return archiveErr
		}
		return fmt.Errorf("parse tribe details: %w", err)
	}
	if err := printArchiveSummary(out, "Tribe save", fileSummary.Archive, opts); err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Tribe name: %s\nTribe ID: %v\nOwner ID: %v\nMembers: %d\nDinos: %d\n",
		displayString(fileSummary.Summary.Name, opts),
		displayInt(fileSummary.Summary.TribeID, opts),
		displayInt(fileSummary.Summary.OwnerID, opts),
		len(fileSummary.Summary.Members),
		fileSummary.Summary.NumDinos,
	)
	return err
}

func tribesDirectory(path string, out io.Writer, opts runOptions) error {
	summary, err := arkapi.TribeDirectorySummaryFromPath(path)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(
		out,
		"Tribe directory: %s\nTribe files: %d\nTribes: %d\nTotal members: %d\nAverage members: %.2f\nTotal dinos: %d\nAverage dinos: %.2f\n",
		displayString(path, opts),
		summary.Files,
		len(summary.Tribes),
		summary.TotalMembers,
		optionalFloat(summary.AverageMembers, summary.HasAverageMembers),
		summary.TotalDinos,
		optionalFloat(summary.AverageDinos, summary.HasAverageDinos),
	); err != nil {
		return err
	}
	if opts.Redact {
		return nil
	}
	for _, tribe := range summary.Tribes {
		if _, err := fmt.Fprintf(
			out,
			"  tribe id=%d name=%s owner=%d members=%d dinos=%d\n",
			tribe.TribeID,
			tribe.Name,
			tribe.OwnerID,
			len(tribe.Members),
			tribe.NumDinos,
		); err != nil {
			return err
		}
	}
	return nil
}

func mutate(args []string, out io.Writer, opts runOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("mutate requires a subcommand")
	}
	switch args[0] {
	case "copy":
		if len(args) != 3 {
			return fmt.Errorf("mutate copy requires a local .ark path and explicit output path")
		}
		if err := arkmutation.CopySave(args[1], args[2]); err != nil {
			return err
		}
		_, err := fmt.Fprintf(out, "Wrote experimental live-server-unverified mutation copy: %s\n", displayString(args[2], opts))
		return err
	case "remove-object":
		if len(args) != 4 {
			return fmt.Errorf("mutate remove-object requires a local .ark path, explicit output path, and object UUID")
		}
		id, err := uuid.Parse(args[3])
		if err != nil {
			return err
		}
		if err := arkmutation.RemoveObject(args[1], args[2], id); err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "Wrote experimental live-server-unverified mutation copy without object %s: %s\n", displayString(id.String(), opts), displayString(args[2], opts))
		return err
	case "remove-class-contains":
		if len(args) != 4 {
			return fmt.Errorf("mutate remove-class-contains requires a local .ark path, explicit output path, and class substring")
		}
		removed, err := arkmutation.RemoveObjectsByClassContains(args[1], args[2], args[3])
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "Wrote experimental live-server-unverified mutation copy without class substring %s removed=%d: %s\n", displayString(args[3], opts), removed, displayString(args[2], opts))
		return err
	case "put-object-hex":
		if len(args) != 5 {
			return fmt.Errorf("mutate put-object-hex requires a local .ark path, explicit output path, object UUID, and hex value")
		}
		id, err := uuid.Parse(args[3])
		if err != nil {
			return err
		}
		value, err := hex.DecodeString(args[4])
		if err != nil {
			return fmt.Errorf("decode object hex value: %w", err)
		}
		if err := arkmutation.PutObjectBinary(args[1], args[2], id, value); err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "Wrote experimental live-server-unverified mutation copy with object %s: %s\n", displayString(id.String(), opts), displayString(args[2], opts))
		return err
	case "replace-object-property-hex":
		if len(args) != 7 {
			return fmt.Errorf("mutate replace-object-property-hex requires a local .ark path, explicit output path, object UUID, property name, property position, and hex-encoded property")
		}
		id, err := uuid.Parse(args[3])
		if err != nil {
			return err
		}
		rawPosition, err := strconv.ParseInt(args[5], 10, 32)
		if err != nil {
			return fmt.Errorf("parse property position: %w", err)
		}
		value, err := hex.DecodeString(args[6])
		if err != nil {
			return fmt.Errorf("decode property hex value: %w", err)
		}
		if err := arkmutation.ReplaceObjectPropertyBinary(args[1], args[2], id, args[4], int32(rawPosition), value); err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "Wrote experimental live-server-unverified mutation copy with replaced property %s[%d] on object %s: %s\n", displayString(args[4], opts), rawPosition, displayString(id.String(), opts), displayString(args[2], opts))
		return err
	case "import-base-binary":
		if len(args) != 4 {
			return fmt.Errorf("mutate import-base-binary requires a local .ark path, explicit output path, and base export directory")
		}
		imported, err := arkmutation.ImportBaseBinary(args[1], args[2], args[3])
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "Wrote experimental live-server-unverified mutation copy with imported base rows=%d: %s\n", imported, displayString(args[2], opts))
		return err
	case "import-structure-binary":
		if len(args) != 4 {
			return fmt.Errorf("mutate import-structure-binary requires a local .ark path, explicit output path, and structure export directory")
		}
		imported, err := arkmutation.ImportStructureBinary(args[1], args[2], args[3])
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "Wrote experimental live-server-unverified mutation copy with imported structure rows=%d: %s\n", imported, displayString(args[2], opts))
		return err
	case "import-dino-binary":
		if len(args) != 4 {
			return fmt.Errorf("mutate import-dino-binary requires a local .ark path, explicit output path, and dino export directory")
		}
		imported, err := arkmutation.ImportDinoBinary(args[1], args[2], args[3])
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "Wrote experimental live-server-unverified mutation copy with imported dino rows=%d: %s\n", imported, displayString(args[2], opts))
		return err
	case "import-equipment-binary":
		if len(args) != 4 {
			return fmt.Errorf("mutate import-equipment-binary requires a local .ark path, explicit output path, and equipment export directory")
		}
		imported, err := arkmutation.ImportEquipmentBinary(args[1], args[2], args[3])
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "Wrote experimental live-server-unverified mutation copy with imported equipment rows=%d: %s\n", imported, displayString(args[2], opts))
		return err
	case "put-custom":
		if len(args) != 5 {
			return fmt.Errorf("mutate put-custom requires a local .ark path, explicit output path, custom key, and hex value")
		}
		value, err := hex.DecodeString(args[4])
		if err != nil {
			return fmt.Errorf("decode custom hex value: %w", err)
		}
		if err := arkmutation.PutCustomValue(args[1], args[2], args[3], value); err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "Wrote experimental live-server-unverified mutation copy with custom key %s: %s\n", displayString(args[3], opts), displayString(args[2], opts))
		return err
	default:
		return fmt.Errorf("unknown mutate subcommand %q", args[0])
	}
}

func cluster(path string, out io.Writer, opts runOptions) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		directory, err := arkapi.ClusterDirectorySummaryFromPath(path)
		if err != nil {
			return err
		}
		if len(directory.Files) == 0 {
			_, err = fmt.Fprintf(out, "Cluster directory: %s\nFiles: 0\nFile faults: %d\n", displayString(path, opts), len(directory.Faults))
			return err
		}
		summary := directory.Summary
		if _, err := fmt.Fprintf(out, "Cluster directory: %s\nFiles: %d\nFile faults: %d\nObjects: %d\nItems: %d\nDinos: %d\nParse errors: %d\n\n", displayString(path, opts), summary.Files, len(directory.Faults), summary.Objects, summary.Items, summary.Dinos, summary.ParseErrors); err != nil {
			return err
		}
		for i, entry := range directory.Files {
			if i > 0 {
				if _, err := fmt.Fprintln(out); err != nil {
					return err
				}
			}
			if err := printClusterSummary(out, entry, opts); err != nil {
				return err
			}
		}
		return nil
	}
	data, err := arkapi.ClusterSummaryFromPath(path)
	if err != nil {
		return err
	}
	return printClusterSummary(out, data, opts)
}

func clusterSummary(path string, out io.Writer, opts runOptions) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		directory, err := arkapi.ClusterDirectorySummaryFromPath(path)
		if err != nil {
			return err
		}
		summary := directory.Summary
		if _, err := fmt.Fprintf(
			out,
			"Cluster directory: %s\nFiles: %d\nFile faults: %d\nObjects: %d\nItems: %d\nDinos: %d\nParse errors: %d\n",
			displayString(path, opts),
			summary.Files,
			len(directory.Faults),
			summary.Objects,
			summary.Items,
			summary.Dinos,
			summary.ParseErrors,
		); err != nil {
			return err
		}
		return printClusterTypedSummaries(out, summary.ItemSummary, summary.DinoSummary)
	}
	fileSummary, err := arkapi.ClusterSummaryFromPath(path)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(
		out,
		"Cluster file: %s\nArchive version: %d\nObjects: %d\nItems: %d\nDinos: %d\nParse errors: %d\n",
		displayString(fileSummary.Path, opts),
		fileSummary.ArchiveVersion,
		fileSummary.ObjectCount,
		fileSummary.ItemCount,
		fileSummary.DinoCount,
		clusterInfoParseErrorCount(fileSummary),
	); err != nil {
		return err
	}
	return printClusterTypedSummaries(out, clusterInfoItemSummary(fileSummary), clusterInfoDinoSummary(fileSummary))
}

func printClusterTypedSummaries(out io.Writer, items arkapi.ClusterItemSummary, dinos arkapi.ClusterDinoSummary) error {
	_, err := fmt.Fprintf(
		out,
		"Dino item uploads: %d\nEquipment item uploads: %d\nOther item uploads: %d\nSupported item uploads: %d\nUnsupported item uploads: %d\nCrafted item uploads: %d\nTotal item quantity: %d\nAverage item quantity: %.2f\nTotal item rating: %.2f\nAverage item rating: %.2f\nMax item rating: %.1f\nMax item quality: %d\nItems with upload time: %d\nEarliest item upload: %.0f\nLatest item upload: %.0f\nParsed dinos: %d\nDino parse errors: %d\nSupported dino uploads: %d\nUnsupported dino uploads: %d\nDinos with status component: %d\nDinos with AI controller: %d\nDinos with inventory component: %d\nDinos with IDs: %d\nTamed dinos: %d\nFemale dinos: %d\nBaby dinos: %d\nDead dinos: %d\nDinos with stats: %d\nTotal base level: %d\nAverage base level: %.2f\nMax base level: %d\nTotal current level: %d\nAverage current level: %.2f\nMax current level: %d\nEmbedded dino objects: %d\nMax embedded dino objects: %d\nDinos with upload time: %d\nEarliest dino upload: %.0f\nLatest dino upload: %.0f\n",
		items.DinoItems,
		items.EquipmentItems,
		items.OtherItems,
		items.SupportedVersionItems,
		items.UnsupportedVersionItems,
		items.CraftedItems,
		items.TotalQuantity,
		items.AverageQuantity,
		items.TotalRating,
		items.AverageRating,
		items.MaxRating,
		items.MaxQuality,
		items.WithUploadTime,
		items.EarliestUploadTime,
		items.LatestUploadTime,
		dinos.ParsedDinos,
		dinos.ParseErrorDinos,
		dinos.SupportedVersionDinos,
		dinos.UnsupportedVersionDinos,
		dinos.WithStatusComponent,
		dinos.WithAIController,
		dinos.WithInventoryComponent,
		dinos.WithDinoID,
		dinos.TamedDinos,
		dinos.FemaleDinos,
		dinos.BabyDinos,
		dinos.DeadDinos,
		dinos.WithStats,
		dinos.TotalBaseLevel,
		dinos.AverageBaseLevel,
		dinos.MaxBaseLevel,
		dinos.TotalCurrentLevel,
		dinos.AverageCurrentLevel,
		dinos.MaxCurrentLevel,
		dinos.TotalEmbeddedObjects,
		dinos.MaxEmbeddedObjects,
		dinos.WithUploadTime,
		dinos.EarliestUploadTime,
		dinos.LatestUploadTime,
	)
	return err
}

func clusterInfoItemSummary(info arkapi.ClusterDataInfo) arkapi.ClusterItemSummary {
	summary := arkapi.ClusterItemSummary{Items: len(info.Items)}
	for _, item := range info.Items {
		switch item.Type {
		case "dino":
			summary.DinoItems++
		case "equipment":
			summary.EquipmentItems++
		default:
			summary.OtherItems++
		}
		if item.SupportedVersion {
			summary.SupportedVersionItems++
		}
		if item.UnsupportedVersion {
			summary.UnsupportedVersionItems++
		}
		if item.IsCrafted {
			summary.CraftedItems++
		}
		summary.TotalQuantity += int64(item.Quantity)
		summary.TotalRating += item.Rating
		if item.Rating > summary.MaxRating {
			summary.MaxRating = item.Rating
		}
		if item.Quality > summary.MaxQuality {
			summary.MaxQuality = item.Quality
		}
		addClusterUploadTime(&summary.WithUploadTime, &summary.EarliestUploadTime, &summary.LatestUploadTime, item.UploadTime)
	}
	if summary.Items > 0 {
		summary.AverageQuantity = float64(summary.TotalQuantity) / float64(summary.Items)
		summary.AverageRating = summary.TotalRating / float64(summary.Items)
	}
	return summary
}

func clusterInfoDinoSummary(info arkapi.ClusterDataInfo) arkapi.ClusterDinoSummary {
	summary := arkapi.ClusterDinoSummary{Dinos: len(info.Dinos)}
	for _, dino := range info.Dinos {
		switch dino.ParseStatus {
		case "parsed":
			summary.ParsedDinos++
		case "parse_error":
			summary.ParseErrorDinos++
		}
		if dino.ParseStatus != "parse_error" && dino.ParseError != "" {
			summary.ParseErrorDinos++
		}
		if dino.SupportedVersion {
			summary.SupportedVersionDinos++
		}
		if dino.UnsupportedVersion {
			summary.UnsupportedVersionDinos++
		}
		if len(dino.StatusComponentClassNames) > 0 {
			summary.WithStatusComponent++
		}
		if len(dino.AIControllerClassNames) > 0 {
			summary.WithAIController++
		}
		if len(dino.InventoryComponentClassNames) > 0 {
			summary.WithInventoryComponent++
		}
		if dino.DinoID1 != 0 || dino.DinoID2 != 0 {
			summary.WithDinoID++
		}
		if dino.IsTamed {
			summary.TamedDinos++
		}
		if dino.IsFemale {
			summary.FemaleDinos++
		}
		if dino.IsBaby {
			summary.BabyDinos++
		}
		if dino.IsDead {
			summary.DeadDinos++
		}
		if dino.HasStats {
			summary.WithStats++
			summary.TotalBaseLevel += int64(dino.BaseLevel)
			if dino.BaseLevel > summary.MaxBaseLevel {
				summary.MaxBaseLevel = dino.BaseLevel
			}
			summary.TotalCurrentLevel += int64(dino.CurrentLevel)
			if dino.CurrentLevel > summary.MaxCurrentLevel {
				summary.MaxCurrentLevel = dino.CurrentLevel
			}
		}
		summary.TotalEmbeddedObjects += dino.ObjectCount
		if dino.ObjectCount > summary.MaxEmbeddedObjects {
			summary.MaxEmbeddedObjects = dino.ObjectCount
		}
		addClusterUploadTime(&summary.WithUploadTime, &summary.EarliestUploadTime, &summary.LatestUploadTime, dino.UploadTime)
	}
	if summary.WithStats > 0 {
		summary.AverageBaseLevel = float64(summary.TotalBaseLevel) / float64(summary.WithStats)
		summary.AverageCurrentLevel = float64(summary.TotalCurrentLevel) / float64(summary.WithStats)
	}
	return summary
}

func addClusterUploadTime(count *int, earliest *float64, latest *float64, uploadTime float64) {
	if uploadTime == 0 {
		return
	}
	if *count == 0 || uploadTime < *earliest {
		*earliest = uploadTime
	}
	if *count == 0 || uploadTime > *latest {
		*latest = uploadTime
	}
	*count++
}

func clusterInfoParseErrorCount(info arkapi.ClusterDataInfo) int {
	return clusterInfoDinoSummary(info).ParseErrorDinos
}

func clusterInfoDinoParseStatusCounts(info arkapi.ClusterDataInfo) map[string]int {
	counts := map[string]int{
		"parsed":              0,
		"unsupported_version": 0,
		"parse_error":         0,
		"unparsed":            0,
	}
	for _, dino := range info.Dinos {
		status := dino.ParseStatus
		if status == "" {
			status = "unparsed"
		}
		if _, ok := counts[status]; !ok {
			counts[status] = 0
		}
		counts[status]++
	}
	return counts
}

func tribute(path string, out io.Writer, opts runOptions) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		directory, err := arkapi.TributeDirectorySummaryFromPath(path)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(out, "Tribute directory: %s\nFiles: %d\nFile faults: %d\n", displayString(path, opts), len(directory.Files), len(directory.Faults)); err != nil {
			return err
		}
		if len(directory.Files) == 0 {
			return nil
		}
		if _, err := fmt.Fprintln(out); err != nil {
			return err
		}
		for i, entry := range directory.Files {
			if i > 0 {
				if _, err := fmt.Fprintln(out); err != nil {
					return err
				}
			}
			if err := printTributeSummary(out, entry, opts); err != nil {
				return err
			}
		}
		return nil
	}
	data, err := arkapi.TributeSummaryFromPath(path)
	if err != nil {
		return err
	}
	return printTributeSummary(out, data, opts)
}

func exportJSON(path string, outputPath string, out io.Writer, opts runOptions) error {
	var data []byte
	var err error
	if opts.Redact {
		data, err = arkapi.ExportRedactedSaveInfoJSONFromPath(path)
		if err != nil {
			return err
		}
	} else {
		data, err = arkapi.ExportSaveInfoJSONFromPath(path)
		if err != nil {
			return err
		}
	}
	if err := os.WriteFile(outputPath, append(data, '\n'), 0o600); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Wrote JSON export: %s\n", displayString(outputPath, opts))
	return err
}

func exportDomainJSON(path string, domain string, outputPath string, out io.Writer, opts runOptions) error {
	var data []byte
	var err error
	if opts.Redact {
		data, err = arkapi.ExportRedactedDomainJSONFromPath(path, domain)
		if err != nil {
			return err
		}
	} else {
		data, err = arkapi.ExportDomainJSONFromPath(path, domain)
		if err != nil {
			return err
		}
	}
	if err := os.WriteFile(outputPath, append(data, '\n'), 0o600); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Wrote %s JSON export: %s\n", domain, displayString(outputPath, opts))
	return err
}

func exportClusterJSON(path string, outputPath string, out io.Writer, opts runOptions) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return exportClusterDirectoryJSON(path, outputPath, out, opts)
	}
	var raw []byte
	if opts.Redact {
		info, err := arkapi.ClusterSummaryFromPath(path)
		if err != nil {
			return err
		}
		info.ID = redactedValue
		info.Path = redactedValue
		info.Items = nil
		info.Dinos = nil
		raw, err = json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}
	} else {
		raw, err = arkapi.ExportClusterPathJSON(path)
		if err != nil {
			return err
		}
	}
	if err := os.WriteFile(outputPath, append(raw, '\n'), 0o600); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Wrote cluster JSON export: %s\n", displayString(outputPath, opts))
	return err
}

func exportClusterDirectoryJSON(path string, outputPath string, out io.Writer, opts runOptions) error {
	var raw []byte
	var err error
	if opts.Redact {
		info, err := arkapi.ClusterDirectorySummaryFromPath(path)
		if err != nil {
			return err
		}
		for i := range info.Files {
			info.Files[i].ID = redactedValue
			info.Files[i].Path = redactedValue
			info.Files[i].Items = nil
			info.Files[i].Dinos = nil
		}
		for i := range info.Faults {
			info.Faults[i].Path = redactedValue
		}
		raw, err = json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}
	} else {
		raw, err = arkapi.ExportClusterPathJSON(path)
		if err != nil {
			return err
		}
	}
	if err := os.WriteFile(outputPath, append(raw, '\n'), 0o600); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Wrote cluster JSON export: %s\n", displayString(outputPath, opts))
	return err
}

func exportTributeJSON(path string, outputPath string, out io.Writer, opts runOptions) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return exportTributeDirectoryJSON(path, outputPath, out, opts)
	}
	var raw []byte
	if opts.Redact {
		info, err := arkapi.TributeSummaryFromPath(path)
		if err != nil {
			return err
		}
		info.ID = redactedValue
		info.Path = redactedValue
		info.PlayerDataIDs = nil
		info.TribeDataIDs = nil
		raw, err = json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}
	} else {
		raw, err = arkapi.ExportTributePathJSON(path)
		if err != nil {
			return err
		}
	}
	if err := os.WriteFile(outputPath, append(raw, '\n'), 0o600); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Wrote tribute JSON export: %s\n", displayString(outputPath, opts))
	return err
}

func exportTributeDirectoryJSON(path string, outputPath string, out io.Writer, opts runOptions) error {
	var raw []byte
	var err error
	if opts.Redact {
		info, err := arkapi.TributeDirectorySummaryFromPath(path)
		if err != nil {
			return err
		}
		for i := range info.Files {
			info.Files[i].ID = redactedValue
			info.Files[i].Path = redactedValue
			info.Files[i].PlayerDataIDs = nil
			info.Files[i].TribeDataIDs = nil
		}
		for i := range info.Faults {
			info.Faults[i].Path = redactedValue
		}
		raw, err = json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}
	} else {
		raw, err = arkapi.ExportTributePathJSON(path)
		if err != nil {
			return err
		}
	}
	if err := os.WriteFile(outputPath, append(raw, '\n'), 0o600); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Wrote tribute JSON export: %s\n", displayString(outputPath, opts))
	return err
}

func printClusterSummary(out io.Writer, data arkapi.ClusterDataInfo, opts runOptions) error {
	if _, err := fmt.Fprintf(out, "Cluster file: %s\nArchive version: %d\nObjects: %d\nItems: %d\nDinos: %d\n", displayString(data.Path, opts), data.ArchiveVersion, data.ObjectCount, data.ItemCount, data.DinoCount); err != nil {
		return err
	}
	if len(data.Dinos) > 0 {
		statuses := clusterInfoDinoParseStatusCounts(data)
		if _, err := fmt.Fprintf(
			out,
			"Dino parse statuses: parsed=%d unsupported_version=%d parse_error=%d unparsed=%d\n",
			statuses["parsed"],
			statuses["unsupported_version"],
			statuses["parse_error"],
			statuses["unparsed"],
		); err != nil {
			return err
		}
	}
	if opts.Redact {
		return nil
	}
	for _, item := range data.Items {
		if _, err := fmt.Fprintf(out, "  item[%d] type=%s short=%s blueprint=%s quantity=%d upload=%.0f\n", item.Index, item.Type, item.ShortName, item.Blueprint, item.Quantity, item.UploadTime); err != nil {
			return err
		}
	}
	for _, dino := range data.Dinos {
		classNames := ""
		if len(dino.ClassNames) > 0 {
			classNames = fmt.Sprintf(" class_names=%s", strings.Join(dino.ClassNames, ","))
		}
		primaryClass := ""
		if dino.PrimaryClassName != "" {
			primaryClass = fmt.Sprintf(" primary_class=%s", dino.PrimaryClassName)
		}
		shortName := ""
		if dino.ShortName != "" {
			shortName = fmt.Sprintf(" short=%s", dino.ShortName)
		}
		details := clusterDinoDetailSuffix(dino)
		if dino.ParseError != "" {
			if _, err := fmt.Fprintf(out, "  dino[%d] raw_bytes=%d objects=%d upload=%.0f%s%s%s%s parse_error=%s\n", dino.Index, dino.RawSize, dino.ObjectCount, dino.UploadTime, primaryClass, shortName, classNames, details, dino.ParseError); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(out, "  dino[%d] raw_bytes=%d objects=%d upload=%.0f%s%s%s%s\n", dino.Index, dino.RawSize, dino.ObjectCount, dino.UploadTime, primaryClass, shortName, classNames, details); err != nil {
				return err
			}
		}
	}
	return nil
}

func clusterDinoDetailSuffix(dino arkapi.ClusterDinoInfo) string {
	var parts []string
	if dino.DinoID1 != 0 || dino.DinoID2 != 0 {
		parts = append(parts, fmt.Sprintf("dino_id=%d/%d", dino.DinoID1, dino.DinoID2))
	}
	if dino.TamedName != "" {
		parts = append(parts, fmt.Sprintf("tamed_name=%s", dino.TamedName))
	}
	if dino.IsTamed {
		parts = append(parts, "tamed=true")
	}
	if dino.IsFemale {
		parts = append(parts, "female=true")
	}
	if dino.IsBaby {
		parts = append(parts, "baby=true")
	}
	if dino.IsDead {
		parts = append(parts, "dead=true")
	}
	if dino.HasStats {
		parts = append(parts, fmt.Sprintf("base_level=%d current_level=%d", dino.BaseLevel, dino.CurrentLevel))
	}
	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, " ")
}

func printArchiveSummary(out io.Writer, label string, summary arkapi.LocalArchiveSummary, opts runOptions) error {
	if _, err := fmt.Fprintf(out, "%s: %s\nArchive version: %d\nObjects: %d\nProperty parse errors: %d\n", label, displayString(summary.Path, opts), summary.ArchiveVersion, summary.ObjectCount, summary.PropertyParseErrors); err != nil {
		return err
	}
	if len(summary.ClassNames) == 0 || opts.Redact {
		return nil
	}
	if _, err := fmt.Fprintln(out, "Classes:"); err != nil {
		return err
	}
	for _, className := range summary.ClassNames {
		if _, err := fmt.Fprintf(out, "  %s\n", className); err != nil {
			return err
		}
	}
	return nil
}

func printTributeSummary(out io.Writer, data arkapi.TributeDataInfo, opts runOptions) error {
	if _, err := fmt.Fprintf(out, "Tribute file: %s\nPlayer data IDs: %d\nTribe data IDs: %d\n", displayString(data.Path, opts), data.PlayerDataCount, data.TribeDataCount); err != nil {
		return err
	}
	if opts.Redact {
		return nil
	}
	for _, id := range data.PlayerDataIDs {
		if _, err := fmt.Fprintf(out, "  player_data_id=%d\n", id); err != nil {
			return err
		}
	}
	for _, id := range data.TribeDataIDs {
		if _, err := fmt.Fprintf(out, "  tribe_data_id=%d\n", id); err != nil {
			return err
		}
	}
	return nil
}

func displayString(value string, opts runOptions) string {
	if opts.Redact {
		return redactedValue
	}
	return value
}

func displayInt[T ~int | ~int16 | ~int32 | ~int64 | ~uint | ~uint16 | ~uint32 | ~uint64](value T, opts runOptions) any {
	if opts.Redact {
		return redactedValue
	}
	return value
}

func optionalFloat(value float64, ok bool) float64 {
	if !ok {
		return 0
	}
	return value
}
