package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <local-cluster-file>", os.Args[0])
	}
	cluster, err := arkcluster.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	api := arkapi.NewCluster(cluster)
	summary := api.Summary()
	items := api.ItemSummary()
	dinos := api.DinoSummary()
	fmt.Printf("cluster=%s items=%d dinos=%d equipment=%d dino_items=%d other_items=%d crafted=%d unsupported_items=%d parsed_dinos=%d unsupported_dinos=%d dino_parse_errors=%d embedded_objects=%d parse_errors=%d\n",
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
		dinos.TotalEmbeddedObjects,
		summary.ParseErrorCount,
	)
}
