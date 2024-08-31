package s3backup

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"s3backup/internal/config"
	"s3backup/internal/export"
	"s3backup/internal/walker"
)

func Main() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})).With("source_dir", cfg.SourceDir, "bucket", cfg.DestinationBucket)
	logger.Info("starting")

	//s3destination, err := export.NewS3Destination(ctx, cfg.DestinationBucket)
	//if err != nil {
	//	return err
	//}

	// exporter := export.NewExporter(s3destination)
	exporter := export.NewExporter(export.NewNoopDestination(logger))

	return walker.EnumerateTopLevelFolders(cfg.SourceDir, func(folder string) error {
		logger := logger.With("folder", folder)
		logger.Info("starting folder upload")

		export, err := walker.SelectFiles(folder)
		if err != nil {
			return fmt.Errorf("preparing export: %w", err)
		}

		if err := exporter.Upload(ctx, logger, export); err != nil {
			return fmt.Errorf("uploading %s: %w", folder, err)
		}
		logger.Info("done")
		return nil
	})
}
