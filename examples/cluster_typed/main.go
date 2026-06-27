package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
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
	fmt.Printf("cluster=%s items=%d dinos=%d equipment=%d dino_items=%d parse_errors=%d\n",
		summary.ID,
		summary.ItemCount,
		summary.DinoCount,
		len(api.ItemsByTypedType(arkobject.ClusterItemTypeEquipment)),
		len(api.ItemsByTypedType(arkobject.ClusterItemTypeDino)),
		summary.ParseErrorCount,
	)
}
