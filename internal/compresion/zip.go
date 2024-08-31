package compresion

import (
	"archive/zip"
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"log/slog"
	"os"
	"s3backup/internal/walker"
)

// Pipe converts all sourceFiles into a zipped stream that should be handled by destination func
func Pipe(ctx context.Context, logger *slog.Logger, progress *walker.Progress, files []walker.ExportFile, destination func(contents io.Reader) error) error {
	pipeReader, pipeWriter := io.Pipe()

	errGroup, ctx := errgroup.WithContext(ctx)

	// upload
	errGroup.Go(func() error {
		err := destination(pipeReader)
		if err != nil {
			return fmt.Errorf("writing to destination: %w", err)
		}

		return pipeReader.Close()
	})

	// zip and write
	errGroup.Go(func() error {
		zipWriter := zip.NewWriter(pipeWriter)
		for _, sourceFile := range files {
			progress.Starting(sourceFile)
			err := zipFile(ctx, logger, sourceFile, zipWriter)
			if err != nil {
				return fmt.Errorf("compressing folder: %w", err)
			}
			progress.Finished(sourceFile)
		}
		progress.Done()
		logger.Info("finished zipping")
		if err := zipWriter.Close(); err != nil {
			return fmt.Errorf("closing zip file: %w", err)
		}
		if err := pipeWriter.Close(); err != nil {
			return err
		}
		logger.Info("writer closed")

		return nil
	})

	return errGroup.Wait()
}

func zipFile(ctx context.Context, logger *slog.Logger, file walker.ExportFile, zipWriter *zip.Writer) error {
	info, err := os.Stat(file.Path)
	if err != nil {
		return fmt.Errorf("lstat %s: %w", file.Path, err)
	}

	fileReader, err := os.Open(file.Path)
	if err != nil {
		return err
	}
	defer func(fileReader *os.File) {
		err := fileReader.Close()
		if err != nil {
			logger.Error("file close failed", "error", err)
		}
	}(fileReader)

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = file.RelPath

	fileWriter, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	copyDone := make(chan error)
	go func() {
		_, err = io.Copy(fileWriter, fileReader)

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
}
