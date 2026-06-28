package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatalf("usage: %s <save.ark> <dino-blueprint> <stat>", os.Args[0])
	}
	stat, ok := arkobject.DinoStatFromString(os.Args[3])
	if !ok {
		log.Fatalf("unknown stat %q", os.Args[3])
	}
	summary, _, err := arkapi.DinoBestStatSummaryFromPath(os.Args[1], arkapi.DinoBestStatOptions{
		Blueprints:      []string{os.Args[2]},
		Stats:           []arkobject.DinoStat{stat},
		OnlyTamed:       true,
		ExcludeCryopods: true,
		BaseStat:        true,
	})
	if err != nil {
		log.Fatal(err)
	}
	if !summary.Found {
		fmt.Println("has_result=0")
		return
	}
	fmt.Printf("has_result=1 stat=%s points=%d level=%d\n", summary.Stat.String(), summary.Points, summary.Level)
}
