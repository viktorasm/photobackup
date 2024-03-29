package main

import (
	"context"
	"fmt"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := Main(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

var k = koanf.New(".")

func Main() error {
	ctx := context.Background()

	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	baseFolder, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current working dir: %w", err)
	}
	logger := slog.With("source_dir", baseFolder, "bucket", cfg.DestinationBucket)
	logger.Info("starting")

	uploader, err := newUploader(ctx, cfg.DestinationBucket)
	if err != nil {
		return err
	}

	return enumerateTopLevelFolders(baseFolder, func(folder string) error {
		logger := logger.With("folder", folder)
		logger.Info("starting folder upload")

		err := uploader.pipeFolder(ctx, logger, folder)
		if err != nil {
			return fmt.Errorf("uploading %s: %w", folder, err)
		}
		logger.Info("done")
		return nil
	})
}

type AppConfig struct {
	DestinationBucket string   `yaml:"destinationBucket"`
	Includes          []string `yaml:"includes"`
}

func loadConfig() (*AppConfig, error) {
	// Load JSON config.
	if err := k.Load(file.Provider("backup-config.yaml"), yaml.Parser()); err != nil {
		return nil, err
	}

	c := AppConfig{}
	if err := k.Unmarshal("", &c); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if c.DestinationBucket == "" {
		return nil, fmt.Errorf("destination bucket is not set")
	}

	return &c, nil
}

func enumerateTopLevelFolders(baseFolder string, handler func(string) error) error {
	entries, err := os.ReadDir(baseFolder)
	if err != nil {
		return err
	}
	for _, info := range entries {
		if !info.IsDir() {
			continue
		}

		// TODO: convert to includes
		if !strings.Contains(info.Name(), "2023") {
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
