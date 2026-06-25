package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arklog"
)

func main() {
	logger := arklog.New(os.Stdout)

	logger.SetLevel(arklog.API, true)
	fmt.Println("API logging enabled.")
	logger.Info("This is an info log.")
	logger.Error("This is an error log.")
	logger.API("This is an API log.")
	logger.Save("This is a save log.")

	logger.SetLevel(arklog.API, false)
	logger.SetLevel(arklog.Error, true)
	fmt.Println()
	fmt.Println("Error log level enabled.")
	logger.Info("This is an info log.")
	logger.Error("This is an error log.")
	logger.API("This is an API log.")
	logger.Save("This is a save log.")

	logger.SetLevel(arklog.All, true)
	fmt.Println()
	fmt.Println("All log levels enabled.")
	logger.Info("This is an info log.")
	logger.Error("This is an error log.")
	logger.API("This is an API log.")
	logger.Save("This is a save log.")
	logger.Debug("This is a debug log.")
	logger.Warning("This is a warning log.")
	logger.Parser("This is a parser log.")
}
