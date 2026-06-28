package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkmutation"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkprofile"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/arktribute"
	"github.com/google/uuid"
)

const redactedValue = "[redacted]"

var ignoredEquipmentNameParts = []string{
	"WeaponCrossbow",
	"WeaponMetalHatchet",
	"WeaponMetalPick",
	"WeaponBow",
	"Chitin",
	"Hide",
	"WeaponPike",
	"WeaponGun",
	"Cloth",
}

type runOptions struct {
	Redact bool
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
	if len(args) == 0 {
		return usage(out)
	}
	switch args[0] {
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
	for _, arg := range args {
		switch arg {
		case "--redact":
			opts.Redact = true
		default:
			if strings.HasPrefix(arg, "--") {
				return opts, nil, fmt.Errorf("unknown option %q", arg)
			}
			filtered = append(filtered, arg)
		}
	}
	return opts, filtered, nil
}

func usage(out io.Writer) error {
	_, err := fmt.Fprintln(out, `Usage:
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
  arksave structure-heatmap <save.ark> <out.json> [resolution] [min-in-cell]
  arksave base-components <save.ark>
  arksave dinos <save.ark>
  arksave dino-wild-tamables <save.ark>
  arksave dino-babies <save.ark>
  arksave dino-best-stat <save.ark>
  arksave dino-most-mutated <save.ark>
  arksave dino-wild-tamed <save.ark>
  arksave equipment-summary <save.ark>
  arksave equipment-saddles <save.ark>
  arksave equipment-best <save.ark>
  arksave equipment-rank <save.ark>
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	ids, err := save.ObjectIDs()
	if err != nil {
		return err
	}
	objects, faults, err := arkapi.NewGeneral(save).ObjectsWithFaults()
	if err != nil {
		return err
	}
	if save.Context == nil {
		return errors.New("save context is nil")
	}
	_, err = fmt.Fprintf(
		out,
		"Save: %s\nMap: %s\nSave version: %d\nObjects: %d\nParsed objects: %d\nParse faults: %d\n",
		displayString(path, opts),
		save.Context.MapName,
		save.Context.SaveVersion,
		len(ids),
		len(objects),
		len(faults),
	)
	return err
}

func mapSummary(path string, out io.Writer) error {
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	info, err := arkapi.NewJSON(save).ExportSaveInfo()
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	classes, err := arkapi.NewGeneral(save).Classes()
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	summary, err := arkapi.NewGeneral(save).ObjectSummary(objectID)
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	summary, err := arkapi.NewGeneral(save).PropertyPositionSummary(objectID)
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	summary, faults, err := arkapi.NewGeneral(save).ClassLookupSummaryWithFaults(classSubstrings)
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	summary, faults, err := arkapi.NewGeneral(save).ClassPropertySummaryWithFaults(classSubstring)
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	summary, err := arkapi.NewGeneral(save).PropertyFilterSummary(propertyNames)
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	summary, faults, err := arkapi.NewStructure(save).HealthSummaryWithFaults()
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	summary, faults, err := arkapi.NewStructure(save).TribeOwnershipSummaryWithFaults(int32(tribeID64))
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	summary, faults, err := arkapi.NewStructure(save).OwnerSummaryWithFaults()
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	export, _, err := arkapi.NewStructure(save).OwnerLocationsWithFaults(mapName, digits, arkapi.NewPlayer(save))
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

func structureHeatmap(path string, outPath string, resolution int, minInCell int, out io.Writer) error {
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	mapName := ""
	if save.Context != nil {
		mapName = save.Context.MapName
	}
	summary, _, err := arkapi.NewStructure(save).SelectedHeatmapSummaryWithFaults(arkapi.StructureHeatmapOptions{
		MapName:      mapName,
		Resolution:   resolution,
		MinInSection: minInCell,
	})
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Cells: %d\nTotal: %d\nMax: %d\nParse faults: %d\nWrote: %s\n",
		summary.NonzeroCells,
		summary.Total,
		summary.Max,
		summary.Faults,
		outPath,
	)
	return err
}

func baseComponents(path string, out io.Writer) error {
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	stats, err := arkapi.NewBase(save, "").ComponentStats()
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	summary, faults, err := arkapi.NewDino(save).PopulationSummaryWithFaults(true)
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	summary, faults, err := arkapi.NewDino(save).WildTamableSummaryWithFaults()
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	summary, faults, err := arkapi.NewDino(save).BabySummaryWithFaults(arkapi.BabyFilterOptions{
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	_, dino, stat, points, ok, faults, err := arkapi.NewDino(save).BestDinoForStatFilteredWithFaults(arkapi.DinoBestStatOptions{})
	if err != nil {
		return err
	}
	if !ok {
		_, err = fmt.Fprintf(out, "Best stat: none\nParse faults: %d\n", len(faults))
		return err
	}
	level := int32(0)
	if dino.Stats != nil {
		level = dino.Stats.CurrentLevel
	}
	_, err = fmt.Fprintf(
		out,
		"Best stat: %s\nPoints: %d\nLevel: %d\nBlueprint: %s\nParse faults: %d\n",
		stat.String(),
		points,
		level,
		dino.ShortName(),
		len(faults),
	)
	return err
}

func dinoMostMutated(path string, out io.Writer) error {
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	_, dino, total, ok, err := arkapi.NewDino(save).MostMutatedTamed()
	if err != nil {
		return err
	}
	if !ok {
		_, err = fmt.Fprintln(out, "Most mutated: none")
		return err
	}
	level := int32(0)
	if dino.Stats != nil {
		level = dino.Stats.CurrentLevel
	}
	_, err = fmt.Fprintf(
		out,
		"Most mutated: %s\nTotal mutation points: %d\nMutation pairs: %d\nLevel: %d\n",
		dino.ShortName(),
		total,
		total/2,
		level,
	)
	return err
}

func dinoWildTamed(path string, out io.Writer) error {
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	api := arkapi.NewDino(save)
	dinos, faults, err := api.WildTamedWithFaults()
	if err != nil {
		return err
	}
	maxLevel := int32(0)
	if level, ok := api.MaxCurrentLevel(dinos); ok {
		maxLevel = level
	}
	_, err = fmt.Fprintf(
		out,
		"Wild-tamed dinos: %d\nMax level: %d\nParse faults: %d\n",
		len(dinos),
		maxLevel,
		len(faults),
	)
	return err
}

func equipmentSummary(path string, out io.Writer) error {
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	summary, faults, err := arkapi.NewEquipment(save).SummaryWithFaults(arkapi.EquipmentFilterOptions{})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Items: %d\nTotal quantity: %d\nWeapon items: %d\nArmor items: %d\nSaddle items: %d\nShield items: %d\nBlueprints: %d\nEquipped: %d\nCrafted: %d\nWith custom data: %d\nCustom data entries: %d\nClasses: %d\nMax quality: %d\nMax rating: %.1f\nMax damage: %.1f\nMax armor: %.1f\nMax durability: %.1f\nParse faults: %d\n",
		summary.Items,
		summary.TotalQuantity,
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
		summary.MaxRating,
		summary.MaxDamage,
		summary.MaxArmor,
		summary.MaxCurrentDurability,
		len(faults),
	)
	return err
}

func equipmentSaddles(path string, out io.Writer) error {
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	summary, faults, err := arkapi.NewEquipment(save).SaddleSummaryWithFaults()
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	api := arkapi.NewEquipment(save)
	_, weapon, weaponOK, weaponFaults, err := api.BestWeaponDamageWithFaults(arkapi.EquipmentFilterOptions{
		Kinds:        []arkobject.EquipmentKind{arkobject.EquipmentWeapon},
		Blueprints:   arkapi.UpstreamWeaponBlueprints(),
		NoBlueprints: true,
	})
	if err != nil {
		return err
	}
	_, armor, armorOK, armorFaults, err := api.BestActualDurabilityWithFaults(arkapi.EquipmentFilterOptions{
		Kinds:        []arkobject.EquipmentKind{arkobject.EquipmentArmor},
		Blueprints:   arkapi.UpstreamArmorBlueprints(),
		NoBlueprints: true,
	})
	if err != nil {
		return err
	}
	if weaponOK {
		if _, err := fmt.Fprintf(
			out,
			"Best weapon damage: %.1f\nBest weapon: %s\nBest weapon crafted: %t\n",
			weapon.Stats.Damage,
			arkobject.ShortNameFromBlueprint(weapon.Blueprint),
			weapon.IsCrafted(),
		); err != nil {
			return err
		}
	} else if _, err := fmt.Fprintln(out, "Best weapon: none"); err != nil {
		return err
	}
	if armorOK {
		if _, err := fmt.Fprintf(
			out,
			"Best armor durability: %.1f\nBest armor: %s\nBest armor crafted: %t\n",
			armor.Stats.Durability,
			arkobject.ShortNameFromBlueprint(armor.Blueprint),
			armor.IsCrafted(),
		); err != nil {
			return err
		}
	} else if _, err := fmt.Fprintln(out, "Best armor: none"); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Parse faults: %d\n", len(weaponFaults)+len(armorFaults))
	return err
}

func equipmentRank(path string, out io.Writer) error {
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	api := arkapi.NewEquipment(save)
	items, faults, err := api.RankedCandidatesWithFaults()
	if err != nil {
		return err
	}
	stats := api.RankStats(items, arkapi.EquipmentRankOptions{
		MinRating:        3,
		ExcludeCrafted:   true,
		IgnoredNameParts: ignoredEquipmentNameParts,
	})
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

func equipmentOwnedBy(path string, blueprint string, tribeIDArg string, out io.Writer, opts runOptions) error {
	tribeID64, err := strconv.ParseInt(tribeIDArg, 10, 32)
	if err != nil {
		return fmt.Errorf("parse tribe id: %w", err)
	}
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	summary, faults, err := arkapi.NewEquipment(save).OwnedSummaryWithFaults(arkapi.EquipmentFilterOptions{
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	api := arkapi.NewStackable(save)
	items, faults, err := api.AllStackablesWithFaults()
	if err != nil {
		return err
	}
	summary := api.StackableSummaryForItems(items)
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	summary, err := arkapi.NewStackable(save).ByClassOwnedSummary([]string{blueprint}, arkobject.ObjectOwner{TribeID: int32(tribeID64)})
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
	api, closeAPI, err := arkapi.NewPlayerFromPath(path, arkapi.PlayerPathOptions{Fallback: arkapi.PlayerPathFallbackPlayers})
	if err != nil {
		return err
	}
	defer closeAPI()

	summary, err := api.PlayerRosterSummary()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Players: %d\nWith names: %d\nHighest level: %d\n",
		summary.Players,
		summary.WithNames,
		summary.HighestLevel,
	)
	return err
}

func tribeRoster(path string, out io.Writer) error {
	api, closeAPI, err := arkapi.NewPlayerFromPath(path, arkapi.PlayerPathOptions{Fallback: arkapi.PlayerPathFallbackTribes})
	if err != nil {
		return err
	}
	defer closeAPI()

	summary, err := api.TribeRosterSummary()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		out,
		"Tribes: %d\nWith names: %d\nMembers: %d\nDinos: %d\n",
		summary.Tribes,
		summary.WithNames,
		summary.Members,
		summary.Dinos,
	)
	return err
}

func playerTribeLinks(path string, out io.Writer) error {
	api, closeAPI, err := arkapi.NewPlayerFromPath(path, arkapi.PlayerPathOptions{})
	if err != nil {
		return err
	}
	defer closeAPI()

	summary, err := api.TribePlayerRelationSummary()
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

	profile, err := arkprofile.OpenPlayerProfile(path)
	if err != nil {
		return err
	}
	if err := printArchiveSummary(out, "Player profile", profile.Path, profile.Archive.Version, profile.Archive.Objects, opts); err != nil {
		return err
	}
	player, err := profile.Player()
	if err != nil {
		return fmt.Errorf("parse player profile details: %w", err)
	}
	_, err = fmt.Fprintf(
		out,
		"Character name: %s\nPlayer name: %s\nPlayer data ID: %v\nTribe ID: %v\nDeaths: %d\n",
		displayString(player.CharacterName, opts),
		displayString(player.PlayerName, opts),
		displayInt(player.PlayerDataID, opts),
		displayInt(player.TribeID, opts),
		player.NumDeaths,
	)
	return err
}

func playersDirectory(path string, out io.Writer, opts runOptions) error {
	api, err := arkapi.NewPlayerFromDirectory(path)
	if err != nil {
		return err
	}
	players, err := api.Players()
	if err != nil {
		return err
	}
	totalDeaths, err := api.TotalDeaths()
	if err != nil {
		return err
	}
	averageDeaths, hasAverageDeaths, err := api.AverageDeaths()
	if err != nil {
		return err
	}
	totalLevel, err := api.TotalLevel()
	if err != nil {
		return err
	}
	averageLevel, hasAverageLevel, err := api.AverageLevel()
	if err != nil {
		return err
	}
	totalExperience, err := api.TotalExperience()
	if err != nil {
		return err
	}
	totalEngramPoints, err := api.TotalEngramPoints()
	if err != nil {
		return err
	}
	unlockedEngrams, err := api.UnlockedEngrams()
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(
		out,
		"Player directory: %s\nProfiles: %d\nPlayers: %d\nTotal deaths: %d\nAverage deaths: %.2f\nTotal level: %d\nAverage level: %.2f\nTotal experience: %.2f\nTotal engram points: %d\nUnlocked engrams: %d\n",
		displayString(path, opts),
		len(api.ProfilePaths()),
		len(players),
		totalDeaths,
		optionalFloat(averageDeaths, hasAverageDeaths),
		totalLevel,
		optionalFloat(averageLevel, hasAverageLevel),
		totalExperience,
		totalEngramPoints,
		len(unlockedEngrams),
	); err != nil {
		return err
	}
	if opts.Redact {
		return nil
	}
	for _, player := range players {
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

	tribe, err := arkprofile.OpenTribeSave(path)
	if err != nil {
		return err
	}
	if err := printArchiveSummary(out, "Tribe save", tribe.Path, tribe.Archive.Version, tribe.Archive.Objects, opts); err != nil {
		return err
	}
	summary, err := tribe.Summary()
	if err != nil {
		return fmt.Errorf("parse tribe details: %w", err)
	}
	_, err = fmt.Fprintf(
		out,
		"Tribe name: %s\nTribe ID: %v\nOwner ID: %v\nMembers: %d\nDinos: %d\n",
		displayString(summary.Name, opts),
		displayInt(summary.TribeID, opts),
		displayInt(summary.OwnerID, opts),
		len(summary.Members),
		summary.NumDinos,
	)
	return err
}

func tribesDirectory(path string, out io.Writer, opts runOptions) error {
	api, err := arkapi.NewPlayerFromDirectory(path)
	if err != nil {
		return err
	}
	tribes, err := api.TribeDetails()
	if err != nil {
		return err
	}
	totalDinos, err := api.TotalTribeDinos()
	if err != nil {
		return err
	}
	averageDinos, hasAverageDinos, err := api.AverageTribeDinos()
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(
		out,
		"Tribe directory: %s\nTribe files: %d\nTribes: %d\nTotal dinos: %d\nAverage dinos: %.2f\n",
		displayString(path, opts),
		len(api.TribePaths()),
		len(tribes),
		totalDinos,
		optionalFloat(averageDinos, hasAverageDinos),
	); err != nil {
		return err
	}
	if opts.Redact {
		return nil
	}
	for _, tribe := range tribes {
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
		entries, err := arkcluster.OpenDirectory(path)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			_, err = fmt.Fprintf(out, "Cluster directory: %s\nFiles: 0\n", displayString(path, opts))
			return err
		}
		summary := arkapi.ClusterDirectorySummary(entries)
		if _, err := fmt.Fprintf(out, "Cluster directory: %s\nFiles: %d\nObjects: %d\nItems: %d\nDinos: %d\nParse errors: %d\n\n", displayString(path, opts), summary.Files, summary.Objects, summary.Items, summary.Dinos, summary.ParseErrors); err != nil {
			return err
		}
		for i, entry := range entries {
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
	data, err := arkcluster.Open(path)
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
		entries, err := arkcluster.OpenDirectory(path)
		if err != nil {
			return err
		}
		summary := arkapi.ClusterDirectorySummary(entries)
		if _, err := fmt.Fprintf(
			out,
			"Cluster directory: %s\nFiles: %d\nObjects: %d\nItems: %d\nDinos: %d\nParse errors: %d\n",
			displayString(path, opts),
			summary.Files,
			summary.Objects,
			summary.Items,
			summary.Dinos,
			summary.ParseErrors,
		); err != nil {
			return err
		}
		return printClusterTypedSummaries(out, summary.ItemSummary, summary.DinoSummary)
	}
	data, err := arkcluster.Open(path)
	if err != nil {
		return err
	}
	fileSummary := arkapi.NewCluster(data).Summary()
	if _, err := fmt.Fprintf(
		out,
		"Cluster file: %s\nArchive version: %d\nObjects: %d\nItems: %d\nDinos: %d\nParse errors: %d\n",
		displayString(data.Path, opts),
		fileSummary.ArchiveVersion,
		fileSummary.ObjectCount,
		fileSummary.ItemCount,
		fileSummary.DinoCount,
		fileSummary.ParseErrorCount,
	); err != nil {
		return err
	}
	api := arkapi.NewCluster(data)
	return printClusterTypedSummaries(out, api.ItemSummary(), api.DinoSummary())
}

func printClusterTypedSummaries(out io.Writer, items arkapi.ClusterItemSummary, dinos arkapi.ClusterDinoSummary) error {
	_, err := fmt.Fprintf(
		out,
		"Dino item uploads: %d\nEquipment item uploads: %d\nOther item uploads: %d\nSupported item uploads: %d\nUnsupported item uploads: %d\nCrafted item uploads: %d\nTotal item quantity: %d\nMax item rating: %.1f\nMax item quality: %d\nParsed dinos: %d\nDino parse errors: %d\nSupported dino uploads: %d\nUnsupported dino uploads: %d\nDinos with status component: %d\nDinos with AI controller: %d\nDinos with inventory component: %d\nEmbedded dino objects: %d\nMax embedded dino objects: %d\n",
		items.DinoItems,
		items.EquipmentItems,
		items.OtherItems,
		items.SupportedVersionItems,
		items.UnsupportedVersionItems,
		items.CraftedItems,
		items.TotalQuantity,
		items.MaxRating,
		items.MaxQuality,
		dinos.ParsedDinos,
		dinos.ParseErrorDinos,
		dinos.SupportedVersionDinos,
		dinos.UnsupportedVersionDinos,
		dinos.WithStatusComponent,
		dinos.WithAIController,
		dinos.WithInventoryComponent,
		dinos.TotalEmbeddedObjects,
		dinos.MaxEmbeddedObjects,
	)
	return err
}

func tribute(path string, out io.Writer, opts runOptions) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		entries, err := arktribute.OpenDirectory(path)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			_, err = fmt.Fprintf(out, "Tribute directory: %s\nFiles: 0\n", displayString(path, opts))
			return err
		}
		for i, entry := range entries {
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
	data, err := arktribute.Open(path)
	if err != nil {
		return err
	}
	return printTributeSummary(out, data, opts)
}

func exportJSON(path string, outputPath string, out io.Writer, opts runOptions) error {
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	api := arkapi.NewJSON(save)
	var data []byte
	if opts.Redact {
		info, err := api.ExportSaveInfo()
		if err != nil {
			return err
		}
		info.Objects = nil
		data, err = json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}
	} else {
		data, err = api.ExportSaveInfoJSON()
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
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	api := arkapi.NewJSON(save)
	var data []byte
	if opts.Redact {
		export, err := api.ExportDomain(domain)
		if err != nil {
			return err
		}
		export.Items = nil
		data, err = json.MarshalIndent(export, "", "  ")
		if err != nil {
			return err
		}
	} else {
		data, err = api.ExportDomainJSON(domain)
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

	data, err := arkcluster.Open(path)
	if err != nil {
		return err
	}
	var raw []byte
	if opts.Redact {
		info := arkapi.ExportClusterData(data)
		info.ID = redactedValue
		info.Path = redactedValue
		info.Items = nil
		info.Dinos = nil
		raw, err = json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}
	} else {
		raw, err = arkapi.ExportClusterDataJSON(data)
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
	entries, err := arkcluster.OpenDirectory(path)
	if err != nil {
		return err
	}
	var raw []byte
	if opts.Redact {
		info := arkapi.ExportClusterDirectoryData(entries)
		for i := range info.Files {
			info.Files[i].ID = redactedValue
			info.Files[i].Path = redactedValue
			info.Files[i].Items = nil
			info.Files[i].Dinos = nil
		}
		raw, err = json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}
	} else {
		raw, err = arkapi.ExportClusterDirectoryDataJSON(entries)
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

	data, err := arktribute.Open(path)
	if err != nil {
		return err
	}
	var raw []byte
	if opts.Redact {
		info := arkapi.ExportTributeData(data)
		info.ID = redactedValue
		info.Path = redactedValue
		info.PlayerDataIDs = nil
		info.TribeDataIDs = nil
		raw, err = json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}
	} else {
		raw, err = arkapi.ExportTributeDataJSON(data)
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
	entries, err := arktribute.OpenDirectory(path)
	if err != nil {
		return err
	}
	var raw []byte
	if opts.Redact {
		info := arkapi.ExportTributeDirectoryData(entries)
		for i := range info.Files {
			info.Files[i].ID = redactedValue
			info.Files[i].Path = redactedValue
			info.Files[i].PlayerDataIDs = nil
			info.Files[i].TribeDataIDs = nil
		}
		raw, err = json.MarshalIndent(info, "", "  ")
		if err != nil {
			return err
		}
	} else {
		raw, err = arkapi.ExportTributeDirectoryDataJSON(entries)
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

func printClusterSummary(out io.Writer, data *arkcluster.Data, opts runOptions) error {
	if _, err := fmt.Fprintf(out, "Cluster file: %s\nArchive version: %d\nObjects: %d\nItems: %d\nDinos: %d\n", displayString(data.Path, opts), data.Archive.Version, len(data.Archive.Objects), len(data.Items), len(data.Dinos)); err != nil {
		return err
	}
	if len(data.Dinos) > 0 {
		statuses := arkapi.NewCluster(data).DinoParseStatusCounts()
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
	clusterInfo := arkapi.ExportClusterData(data)
	for _, item := range clusterInfo.Items {
		if _, err := fmt.Fprintf(out, "  item[%d] type=%s short=%s blueprint=%s quantity=%d upload=%.0f\n", item.Index, item.Type, item.ShortName, item.Blueprint, item.Quantity, item.UploadTime); err != nil {
			return err
		}
	}
	for _, dino := range clusterInfo.Dinos {
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
		if dino.ParseError != "" {
			if _, err := fmt.Fprintf(out, "  dino[%d] raw_bytes=%d objects=%d upload=%.0f%s%s%s parse_error=%s\n", dino.Index, dino.RawSize, dino.ObjectCount, dino.UploadTime, primaryClass, shortName, classNames, dino.ParseError); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(out, "  dino[%d] raw_bytes=%d objects=%d upload=%.0f%s%s%s\n", dino.Index, dino.RawSize, dino.ObjectCount, dino.UploadTime, primaryClass, shortName, classNames); err != nil {
				return err
			}
		}
	}
	return nil
}

func printArchiveSummary(out io.Writer, label string, path string, version int32, objects []arkarchive.Object, opts runOptions) error {
	if _, err := fmt.Fprintf(out, "%s: %s\nArchive version: %d\nObjects: %d\nProperty parse errors: %d\n", label, displayString(path, opts), version, len(objects), archivePropertyErrorCount(objects)); err != nil {
		return err
	}
	if len(objects) == 0 || opts.Redact {
		return nil
	}
	if _, err := fmt.Fprintln(out, "Classes:"); err != nil {
		return err
	}
	for _, object := range objects {
		if _, err := fmt.Fprintf(out, "  %s\n", object.ClassName); err != nil {
			return err
		}
	}
	return nil
}

func archivePropertyErrorCount(objects []arkarchive.Object) int {
	var count int
	for _, object := range objects {
		if object.PropertyError != nil {
			count++
		}
	}
	return count
}

func printTributeSummary(out io.Writer, data *arktribute.Data, opts runOptions) error {
	if _, err := fmt.Fprintf(out, "Tribute file: %s\nPlayer data IDs: %d\nTribe data IDs: %d\n", displayString(data.Path, opts), len(data.PlayerDataIDs), len(data.TribeDataIDs)); err != nil {
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
