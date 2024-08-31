package main

import (
	"log/slog"
	"os"
	"s3backup"
)

func main() {
	if err := s3backup.Main(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
