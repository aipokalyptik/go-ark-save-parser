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
	"github.com/aipokalyptik/go-ark-save-parser/arkprofile"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/arktribute"
	"github.com/google/uuid"
)

const redactedValue = "[redacted]"

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
  arksave [--redact] players <player.arkprofile-or-directory>
  arksave [--redact] tribes <tribe.arktribe-or-directory>
  arksave [--redact] cluster <cluster-file-or-directory>
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
	if opts.Redact {
		return nil
	}
	clusterInfo := arkapi.ExportClusterData(data)
	for _, item := range clusterInfo.Items {
		if _, err := fmt.Fprintf(out, "  item[%d] type=%s blueprint=%s quantity=%d upload=%.0f\n", item.Index, item.Type, item.Blueprint, item.Quantity, item.UploadTime); err != nil {
			return err
		}
	}
	for _, dino := range clusterInfo.Dinos {
		classNames := ""
		if len(dino.ClassNames) > 0 {
			classNames = fmt.Sprintf(" class_names=%s", strings.Join(dino.ClassNames, ","))
		}
		if dino.ParseError != "" {
			if _, err := fmt.Fprintf(out, "  dino[%d] raw_bytes=%d objects=%d upload=%.0f%s parse_error=%s\n", dino.Index, dino.RawSize, dino.ObjectCount, dino.UploadTime, classNames, dino.ParseError); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(out, "  dino[%d] raw_bytes=%d objects=%d upload=%.0f%s\n", dino.Index, dino.RawSize, dino.ObjectCount, dino.UploadTime, classNames); err != nil {
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
