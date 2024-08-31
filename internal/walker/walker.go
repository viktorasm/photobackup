package walker

import (
	"crypto/sha256"
	"encoding/hex"
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
	Hash  string
}

type ExportFile struct {
	Path    string
	RelPath string
	Size    int64
}

func (o *ExportObject) sort() {
	slices.SortFunc(o.Files, func(a, b ExportFile) int {
		return strings.Compare(a.RelPath, b.RelPath)
	})
}

func SelectFiles(dir string, excludes []string) (*ExportObject, error) {
	result := ExportObject{
		Name: regexp.MustCompile(`\s+`).ReplaceAllString(filepath.Base(dir), "-"),
	}

	err := filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		excluded, err := matchesPatterns(filePath, excludes)
		if err != nil {
			return err
		}
		if excluded {
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

	result.sort()

	result.Hash = filesHash(result.Files)

	return &result, nil
}

func filesHash(files []ExportFile) string {
	h := sha256.New()
	for _, file := range files {
		h.Write([]byte(fmt.Sprintf("%s:%d", file.RelPath, file.Size)))
	}

	return hex.EncodeToString(h.Sum(nil))
}

func matchesPatterns(path string, patterns []string) (bool, error) {
	path = filepath.ToSlash(path)

	for _, pattern := range patterns {
		match, err := filepath.Match(pattern, path)
		if err != nil {
			return false, fmt.Errorf("matching ignored files: %w", err)
		}
		if match {
			return true, nil
		}
	}
	return false, nil

}

func EnumerateTopLevelFolders(baseFolder string, includes []string, handler func(string) error) error {
	entries, err := os.ReadDir(baseFolder)
	if err != nil {
		return err
	}
	for _, info := range entries {
		if !info.IsDir() {
			continue
		}

		matches, err := matchesPatterns(info.Name(), includes)
		if err != nil {
			return err
		}
		if !matches {
			continue
		}

		if err := handler(filepath.Join(baseFolder, info.Name())); err != nil {
			return err
		}
	}
	return nil
}
