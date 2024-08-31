package export

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"s3backup/internal/compresion"
	"s3backup/internal/walker"
)

type Backend interface {
	Exists(ctx context.Context, path string) (bool, error)
	Write(ctx context.Context, name string, source io.Reader) error
}

type noopBackend struct {
}

func (n noopBackend) Exists(ctx context.Context, path string) (bool, error) {
	return false, nil
}

func (n noopBackend) Write(ctx context.Context, name string, source io.Reader) error {
	slog.Info("noop backend write started", "name", name)
	_, err := io.Copy(io.Discard, source)
	slog.Info("noop backend write finished", "name", name, "err", err)
	return err
}

func NewNoopBackend() Backend {
	return &noopBackend{}
}

var _ Backend = (*noopBackend)(nil)

type Exporter struct {
	b Backend
}

func NewExporter(b Backend) *Exporter {
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

	return compresion.Pipe(ctx, logger, export.Files, func(source io.Reader) error {
		logger.Info("uploading", "object", objectName)

		err := e.b.Write(ctx, objectName, source)
		if err != nil {
			return fmt.Errorf("put object: %w", err)
		}
		return nil
	})
}
