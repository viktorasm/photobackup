package config

import (
	"fmt"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"os"
)

type AppConfig struct {
	DestinationBucket string   `yaml:"destinationBucket"`
	Includes          []string `yaml:"includes"`
	SourceDir         string   `yaml:"sourceDir"`
}

func Load() (*AppConfig, error) {

	var k = koanf.New(".")

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

	if c.SourceDir == "" {
		sourceDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getting current working dir: %w", err)
		}

		c.SourceDir = sourceDir
	}

	return &c, nil
}
