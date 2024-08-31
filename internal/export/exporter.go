package export

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"s3backup/internal/compresion"
	"s3backup/internal/walker"
)

type Destination interface {
	Exists(ctx context.Context, path string) (bool, error)
	Write(ctx context.Context, name string, source io.Reader) error
}

type noopDestination struct {
}

func (n noopDestination) Exists(ctx context.Context, path string) (bool, error) {
	return false, nil
}

func (n noopDestination) Write(ctx context.Context, name string, source io.Reader) error {
	_, err := io.Copy(io.Discard, source)
	return err
}

func NewNoopDestination() Destination {
	return &noopDestination{}
}

type Exporter struct {
	b Destination
}

func NewExporter(b Destination) *Exporter {
	return &Exporter{b}
}

func (e *Exporter) Upload(ctx context.Context, logger *slog.Logger, export *walker.ExportObject) error {
	objectName := fmt.Sprintf("%s.zip", export.Name)

	exists, err := e.b.Exists(ctx, objectName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	progress := walker.NewProgress(objectName, export.Files)

	return compresion.Pipe(ctx, logger, progress, export.Files, func(source io.Reader) error {
		err := e.b.Write(ctx, objectName, source)
		if err != nil {
			return fmt.Errorf("write failed: %w", err)
		}
		return nil
	})
}
