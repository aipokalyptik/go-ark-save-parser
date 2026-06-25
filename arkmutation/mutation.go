package arkmutation

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

var ErrOutputExists = errors.New("mutation output already exists")

func CopySave(inputPath string, outputPath string) error {
	if outputPath == "" {
		return errors.New("mutation output path is required")
	}
	inputAbs, err := filepath.Abs(inputPath)
	if err != nil {
		return err
	}
	outputAbs, err := filepath.Abs(outputPath)
	if err != nil {
		return err
	}
	if inputAbs == outputAbs {
		return errors.New("mutation output path must differ from input path")
	}
	if _, err := os.Stat(outputAbs); err == nil {
		return fmt.Errorf("%w: %s", ErrOutputExists, outputPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	in, err := os.Open(inputAbs)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(outputAbs, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	cleanup := true
	defer func() {
		_ = out.Close()
		if cleanup {
			_ = os.Remove(outputAbs)
		}
	}()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func RemoveObject(inputPath string, outputPath string, id uuid.UUID) error {
	return mutateCopy(inputPath, outputPath, func(db *sql.DB) error {
		_, err := db.Exec(`delete from game where key = ?`, id[:])
		return err
	})
}

func RemoveObjectsByClassContains(inputPath string, outputPath string, substring string) (int, error) {
	if substring == "" {
		return 0, errors.New("class substring is required")
	}
	ids, err := matchingObjectIDsByClassContains(inputPath, substring)
	if err != nil {
		return 0, err
	}
	err = mutateCopy(inputPath, outputPath, func(db *sql.DB) error {
		for _, id := range ids {
			if _, err := db.Exec(`delete from game where key = ?`, id[:]); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(ids), nil
}

func PutObjectBinary(inputPath string, outputPath string, id uuid.UUID, value []byte) error {
	return mutateCopy(inputPath, outputPath, func(db *sql.DB) error {
		_, err := db.Exec(`insert into game (key, value) values (?, ?)
			on conflict(key) do update set value = excluded.value`, id[:], value)
		return err
	})
}

func PutCustomValue(inputPath string, outputPath string, key string, value []byte) error {
	return mutateCopy(inputPath, outputPath, func(db *sql.DB) error {
		_, err := db.Exec(`insert into custom (key, value) values (?, ?)
			on conflict(key) do update set value = excluded.value`, key, value)
		return err
	})
}

func ImportBaseBinary(inputPath string, outputPath string, baseExportDir string) (int, error) {
	rows, err := readExportRows(baseExportDir, []string{"str_"})
	if err != nil {
		return 0, err
	}
	return putObjectRows(inputPath, outputPath, rows)
}

func ImportStructureBinary(inputPath string, outputPath string, structureExportDir string) (int, error) {
	if structureExportDir == "" {
		return 0, errors.New("structure export directory is required")
	}
	if _, err := os.Stat(filepath.Join(structureExportDir, "manifest.json")); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, fmt.Errorf("structure export directory %s is missing manifest.json", structureExportDir)
		}
		return 0, err
	}
	rows, err := readExportRows(structureExportDir, []string{"str_"})
	if err != nil {
		return 0, err
	}
	return putObjectRows(inputPath, outputPath, rows)
}

func ImportDinoBinary(inputPath string, outputPath string, dinoExportDir string) (int, error) {
	rows, err := readExportRows(dinoExportDir, []string{"dino_", "status_", "inv_"})
	if err != nil {
		return 0, err
	}
	return putObjectRows(inputPath, outputPath, rows)
}

func ImportEquipmentBinary(inputPath string, outputPath string, equipmentExportDir string) (int, error) {
	rows, err := readExportRows(equipmentExportDir, []string{"item_"})
	if err != nil {
		return 0, err
	}
	return putObjectRows(inputPath, outputPath, rows)
}

func putObjectRows(inputPath string, outputPath string, rows map[uuid.UUID][]byte) (int, error) {
	err := mutateCopy(inputPath, outputPath, func(db *sql.DB) error {
		for id, value := range rows {
			if _, err := db.Exec(`insert into game (key, value) values (?, ?)
				on conflict(key) do update set value = excluded.value`, id[:], value); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

func readExportRows(exportDir string, prefixes []string) (map[uuid.UUID][]byte, error) {
	if exportDir == "" {
		return nil, errors.New("export directory is required")
	}
	rows := map[uuid.UUID][]byte{}
	err := filepath.WalkDir(exportDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		name := entry.Name()
		prefix, ok := matchingExportPrefix(name, prefixes)
		if !ok || !strings.HasSuffix(name, ".bin") {
			return nil
		}
		rawID := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".bin")
		id, err := uuid.Parse(rawID)
		if err != nil {
			return fmt.Errorf("parse export row filename %s: %w", name, err)
		}
		value, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rows[id] = value
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("export directory %s contains no supported row files", exportDir)
	}
	return rows, nil
}

func matchingExportPrefix(name string, prefixes []string) (string, bool) {
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return prefix, true
		}
	}
	return "", false
}

func matchingObjectIDsByClassContains(inputPath string, substring string) ([]uuid.UUID, error) {
	save, err := arksave.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer save.Close()
	infos, err := save.ObjectClassInfos()
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0)
	for _, info := range infos {
		if strings.Contains(info.ClassName, substring) {
			ids = append(ids, info.UUID)
		}
	}
	return ids, nil
}

func mutateCopy(inputPath string, outputPath string, fn func(*sql.DB) error) error {
	if err := CopySave(inputPath, outputPath); err != nil {
		return err
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(outputPath)
		}
	}()
	db, err := sql.Open("sqlite", outputPath)
	if err != nil {
		return err
	}
	if err := fn(db); err != nil {
		_ = db.Close()
		return err
	}
	if err := db.Close(); err != nil {
		return err
	}
	save, err := arksave.Open(outputPath)
	if err != nil {
		return fmt.Errorf("reopen mutated copy: %w", err)
	}
	if err := save.Close(); err != nil {
		return err
	}
	cleanup = false
	return nil
}
