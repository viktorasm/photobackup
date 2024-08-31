package walker

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type ExportObject struct {
	Name  string
	Files []ExportFile
}
type ExportFile struct {
	Path    string
	RelPath string
	Size    int64
}

func SelectFiles(dir string) (*ExportObject, error) {
	result := ExportObject{
		Name: regexp.MustCompile(`\s+`).ReplaceAllString(filepath.Base(dir), "-"),
	}

	err := filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// TODO: convert to excludes
		if strings.Contains(filePath, "darktable_exported") {
			return nil
		}
		if strings.HasSuffix(filePath, ".xmp") {
			return nil
		}

		slog.Debug("entering", "filepath", filePath)

		if !info.Mode().IsRegular() {
			return nil
		}

		relPath, err := filepath.Rel(dir, filePath)
		if err != nil {
			return err
		}

		relPath = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

		result.Files = append(result.Files, ExportFile{
			Path:    filePath,
			RelPath: relPath,
			Size:    info.Size(),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking folder: %w", err)
	}

	slices.SortFunc(result.Files, func(a, b ExportFile) int {
		return strings.Compare(a.RelPath, b.RelPath)
	})

	return &result, nil
}

func EnumerateTopLevelFolders(baseFolder string, handler func(string) error) error {
	entries, err := os.ReadDir(baseFolder)
	if err != nil {
		return err
	}
	for _, info := range entries {
		if !info.IsDir() {
			continue
		}

		// TODO: convert to includes
		if !strings.Contains(info.Name(), "sicily") {
			continue
		}
		//if !strings.HasSuffix(info.Name(), "09 20") {
		//	return nil
		//}

		if err := handler(filepath.Join(baseFolder, info.Name())); err != nil {
			return err
		}
	}
	return nil
}
