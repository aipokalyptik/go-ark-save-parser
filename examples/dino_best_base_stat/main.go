package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatalf("usage: %s <save.ark> <dino-blueprint> <stat>", os.Args[0])
	}
	stat, ok := arkobject.DinoStatFromString(os.Args[3])
	if !ok {
		log.Fatalf("unknown stat %q", os.Args[3])
	}
	save, err := arksave.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer save.Close()

	api := arkapi.NewDino(save)
	_, dino, gotStat, points, found, _, err := api.BestDinoForStatFilteredWithFaults(arkapi.DinoBestStatOptions{
		Blueprints:      []string{os.Args[2]},
		Stats:           []arkobject.DinoStat{stat},
		OnlyTamed:       true,
		ExcludeCryopods: true,
		BaseStat:        true,
	})
	if err != nil {
		log.Fatal(err)
	}
	if !found {
		fmt.Println("has_result=0")
		return
	}
	level := int32(0)
	if dino.Stats != nil {
		level = dino.Stats.CurrentLevel
	}
	fmt.Printf("has_result=1 stat=%s points=%d level=%d\n", gotStat.String(), points, level)
}
