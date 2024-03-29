package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func zipFolder(ctx context.Context, folderPath string, zipWriter *zip.Writer) error {
	err := filepath.Walk(folderPath, func(filePath string, info os.FileInfo, err error) error {
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

		relPath, err := filepath.Rel(folderPath, filePath)
		if err != nil {
			return err
		}

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath

		fileWriter, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		copyDone := make(chan error)
		go func() {
			slog.Info("adding file to archive", "file", filePath)
			_, err = io.Copy(fileWriter, file)

			copyDone <- err
			close(copyDone)
		}()

		select {
		case <-ctx.Done():
			// Context is canceled, exit the loop
			return ctx.Err()
		case err := <-copyDone:
			return err
		}
	})

	if err != nil {
		return fmt.Errorf("walking folder: %w", err)
	}

	return nil
}
