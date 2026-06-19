package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkmutation"
	"github.com/aipokalyptik/go-ark-save-parser/arkprofile"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return usage(out)
	}
	switch args[0] {
	case "inspect", "parse":
		if len(args) != 2 {
			return fmt.Errorf("%s requires a local .ark path", args[0])
		}
		return inspect(args[1], out)
	case "players":
		if len(args) != 2 {
			return fmt.Errorf("players requires a local .arkprofile path")
		}
		return players(args[1], out)
	case "tribes":
		if len(args) != 2 {
			return fmt.Errorf("tribes requires a local .arktribe path")
		}
		return tribes(args[1], out)
	case "cluster":
		if len(args) != 2 {
			return fmt.Errorf("cluster requires a local cluster file or directory path")
		}
		return cluster(args[1], out)
	case "export-json":
		if len(args) != 3 {
			return fmt.Errorf("export-json requires a local .ark path and explicit output path")
		}
		return exportJSON(args[1], args[2], out)
	case "mutate":
		return mutate(args[1:], out)
	case "ftp", "rcon":
		return fmt.Errorf("%s is unsupported: this is an offline-only parser", args[0])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func usage(out io.Writer) error {
	_, err := fmt.Fprintln(out, `Usage:
  arksave inspect <save.ark>
  arksave parse <save.ark>
  arksave players <player.arkprofile>
  arksave tribes <tribe.arktribe>
  arksave cluster <cluster-file-or-directory>
  arksave export-json <save.ark> <out.json>
  arksave mutate copy <save.ark> <out.ark>
  arksave mutate remove-object <save.ark> <out.ark> <uuid>

Offline-only scope: FTP and RCON are intentionally unsupported.`)
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

func players(path string, out io.Writer) error {
	profile, err := arkprofile.OpenPlayerProfile(path)
	if err != nil {
		return err
	}
	return printArchiveSummary(out, "Player profile", profile.Path, profile.Archive.Version, profile.Archive.Objects)
}

func tribes(path string, out io.Writer) error {
	tribe, err := arkprofile.OpenTribeSave(path)
	if err != nil {
		return err
	}
	if err := printArchiveSummary(out, "Tribe save", tribe.Path, tribe.Archive.Version, tribe.Archive.Objects); err != nil {
		return err
	}
	summary, err := tribe.Summary()
	if err != nil {
		return nil
	}
	_, err = fmt.Fprintf(
		out,
		"Tribe name: %s\nTribe ID: %d\nOwner ID: %d\nMembers: %d\nDinos: %d\n",
		summary.Name,
		summary.TribeID,
		summary.OwnerID,
		len(summary.Members),
		summary.NumDinos,
	)
	return err
}

func mutate(args []string, out io.Writer) error {
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
		_, err := fmt.Fprintf(out, "Wrote mutation copy: %s\n", args[2])
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
		_, err = fmt.Fprintf(out, "Wrote mutation copy without object %s: %s\n", id, args[2])
		return err
	default:
		return fmt.Errorf("unknown mutate subcommand %q", args[0])
	}
}

func cluster(path string, out io.Writer) error {
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
			_, err = fmt.Fprintf(out, "Cluster directory: %s\nFiles: 0\n", path)
			return err
		}
		for i, entry := range entries {
			if i > 0 {
				if _, err := fmt.Fprintln(out); err != nil {
					return err
				}
			}
			if err := printClusterSummary(out, entry); err != nil {
				return err
			}
		}
		return nil
	}
	data, err := arkcluster.Open(path)
	if err != nil {
		return err
	}
	return printClusterSummary(out, data)
}

func exportJSON(path string, outputPath string, out io.Writer) error {
	save, err := arksave.Open(path)
	if err != nil {
		return err
	}
	defer save.Close()

	data, err := arkapi.NewJSON(save).ExportSaveInfoJSON()
	if err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, append(data, '\n'), 0o644); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Wrote JSON export: %s\n", outputPath)
	return err
}

func printClusterSummary(out io.Writer, data *arkcluster.Data) error {
	if _, err := fmt.Fprintf(out, "Cluster file: %s\nArchive version: %d\nObjects: %d\nItems: %d\nDinos: %d\n", data.Path, data.Archive.Version, len(data.Archive.Objects), len(data.Items), len(data.Dinos)); err != nil {
		return err
	}
	for _, item := range data.Items {
		if _, err := fmt.Fprintf(out, "  item[%d] blueprint=%s quantity=%d upload=%.0f\n", item.Index, item.Blueprint, item.Quantity, item.UploadTime); err != nil {
			return err
		}
	}
	for _, dino := range data.Dinos {
		objectCount := 0
		if dino.Archive != nil {
			objectCount = len(dino.Archive.Objects)
		}
		if _, err := fmt.Fprintf(out, "  dino[%d] raw_bytes=%d objects=%d upload=%.0f\n", dino.Index, dino.RawSize, objectCount, dino.UploadTime); err != nil {
			return err
		}
	}
	return nil
}

func printArchiveSummary(out io.Writer, label string, path string, version int32, objects []arkarchive.Object) error {
	if _, err := fmt.Fprintf(out, "%s: %s\nArchive version: %d\nObjects: %d\n", label, path, version, len(objects)); err != nil {
		return err
	}
	if len(objects) == 0 {
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
