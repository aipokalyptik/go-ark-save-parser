package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arksave"
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
	case "players", "tribes", "export-json", "mutate":
		return fmt.Errorf("%s is planned but not implemented yet in this offline port", args[0])
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
