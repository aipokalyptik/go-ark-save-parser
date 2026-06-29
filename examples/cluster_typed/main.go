package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <local-cluster-file>", os.Args[0])
	}
	api, err := arkapi.NewClusterFromPath(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	summary := api.Summary()
	items := api.ItemSummary()
	dinos := api.DinoSummary()
	parseStatuses := api.DinoParseStatusCounts()
	fmt.Printf("cluster=%s items=%d dinos=%d equipment=%d dino_items=%d other_items=%d crafted=%d unsupported_items=%d parsed_dinos=%d unsupported_dinos=%d dino_parse_errors=%d unparsed_dinos=%d dino_ids=%d tamed_dinos=%d female_dinos=%d baby_dinos=%d dead_dinos=%d dinos_with_stats=%d embedded_objects=%d parse_errors=%d\n",
		summary.ID,
		summary.ItemCount,
		summary.DinoCount,
		items.EquipmentItems,
		items.DinoItems,
		items.OtherItems,
		items.CraftedItems,
		items.UnsupportedVersionItems,
		dinos.ParsedDinos,
		dinos.UnsupportedVersionDinos,
		dinos.ParseErrorDinos,
		parseStatuses["unparsed"],
		dinos.WithDinoID,
		dinos.TamedDinos,
		dinos.FemaleDinos,
		dinos.BabyDinos,
		dinos.DeadDinos,
		dinos.WithStats,
		dinos.TotalEmbeddedObjects,
		summary.ParseErrorCount,
	)
}
