package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aipokalyptik/go-ark-save-parser/arkmutation"
	"github.com/google/uuid"
)

func main() {
	if len(os.Args) == 3 {
		copySave(os.Args[1], os.Args[2])
		return
	}
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "copy":
		if len(os.Args) != 4 {
			usage()
		}
		copySave(os.Args[2], os.Args[3])
	case "remove-object":
		if len(os.Args) != 5 {
			usage()
		}
		id := mustUUID(os.Args[4])
		if err := arkmutation.RemoveObject(os.Args[2], os.Args[3], id); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("removed object: %s\nwrote copy: %s\n", id, os.Args[3])
	case "remove-class-contains":
		if len(os.Args) != 5 {
			usage()
		}
		removed, err := arkmutation.RemoveObjectsByClassContains(os.Args[2], os.Args[3], os.Args[4])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("removed class substring: %s removed=%d\nwrote copy: %s\n", os.Args[4], removed, os.Args[3])
	case "put-object-hex":
		if len(os.Args) != 6 {
			usage()
		}
		id := mustUUID(os.Args[4])
		value := mustHex(os.Args[5])
		if err := arkmutation.PutObjectBinary(os.Args[2], os.Args[3], id, value); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("wrote object bytes: %s bytes=%d copy=%s\n", id, len(value), os.Args[3])
	case "replace-object-property-hex":
		if len(os.Args) != 8 {
			usage()
		}
		id := mustUUID(os.Args[4])
		position, err := strconv.ParseInt(os.Args[6], 10, 32)
		if err != nil {
			log.Fatal(err)
		}
		value := mustHex(os.Args[7])
		if err := arkmutation.ReplaceObjectPropertyBinary(os.Args[2], os.Args[3], id, os.Args[5], int32(position), value); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("replaced object property: %s[%d] object=%s bytes=%d copy=%s\n", os.Args[5], position, id, len(value), os.Args[3])
	case "import-base-binary":
		if len(os.Args) != 5 {
			usage()
		}
		imported, err := arkmutation.ImportBaseBinary(os.Args[2], os.Args[3], os.Args[4])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("imported base rows: %d\nwrote copy: %s\n", imported, os.Args[3])
	case "import-structure-binary":
		if len(os.Args) != 5 {
			usage()
		}
		imported, err := arkmutation.ImportStructureBinary(os.Args[2], os.Args[3], os.Args[4])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("imported structure rows: %d\nwrote copy: %s\n", imported, os.Args[3])
	case "import-dino-binary":
		if len(os.Args) != 5 {
			usage()
		}
		imported, err := arkmutation.ImportDinoBinary(os.Args[2], os.Args[3], os.Args[4])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("imported dino rows: %d\nwrote copy: %s\n", imported, os.Args[3])
	case "import-equipment-binary":
		if len(os.Args) != 5 {
			usage()
		}
		imported, err := arkmutation.ImportEquipmentBinary(os.Args[2], os.Args[3], os.Args[4])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("imported equipment rows: %d\nwrote copy: %s\n", imported, os.Args[3])
	case "put-custom":
		if len(os.Args) != 6 {
			usage()
		}
		value := mustHex(os.Args[5])
		if err := arkmutation.PutCustomValue(os.Args[2], os.Args[3], os.Args[4], value); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("wrote custom value: %s bytes=%d copy=%s\n", os.Args[4], len(value), os.Args[3])
	default:
		usage()
	}
}

func copySave(input string, output string) {
	if err := arkmutation.CopySave(input, output); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote copy: %s\n", output)
}

func mustUUID(raw string) uuid.UUID {
	id, err := uuid.Parse(raw)
	if err != nil {
		log.Fatal(err)
	}
	return id
}

func mustHex(raw string) []byte {
	value, err := hex.DecodeString(raw)
	if err != nil {
		log.Fatal(err)
	}
	return value
}

func usage() {
	log.Fatalf(`usage:
  %s <save.ark> <out.ark>
  %s copy <save.ark> <out.ark>
  %s remove-object <save.ark> <out.ark> <uuid>
  %s remove-class-contains <save.ark> <out.ark> <class-substring>
  %s import-base-binary <save.ark> <out.ark> <base-export-dir>
  %s import-structure-binary <save.ark> <out.ark> <structure-export-dir>
  %s import-dino-binary <save.ark> <out.ark> <dino-export-dir>
  %s import-equipment-binary <save.ark> <out.ark> <equipment-export-dir>
  %s put-object-hex <save.ark> <out.ark> <uuid> <hex-value>
  %s replace-object-property-hex <save.ark> <out.ark> <uuid> <property-name> <position> <hex-encoded-property>
  %s put-custom <save.ark> <out.ark> <key> <hex-value>`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}
