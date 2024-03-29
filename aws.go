package main

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"golang.org/x/sync/errgroup"
	"io"
	"log/slog"
	"path/filepath"
	"regexp"
)

type uploader struct {
	client     *s3.Client
	bucketName string
}

func newUploader(ctx context.Context, bucketName string) (*uploader, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating s3 client: %w", err)
	}

	return &uploader{
		bucketName: bucketName,
		client:     s3.NewFromConfig(awsConfig),
	}, nil
}

// UploadLargeObject uses an upload manager to upload data to an object in a bucket.
// The upload manager breaks large data into parts and uploads the parts concurrently.
func (u *uploader) uploadLargeObject(ctx context.Context, objectKey string, contents io.Reader) error {
	var partMiBs int64 = 10
	uploader := manager.NewUploader(u.client, func(u *manager.Uploader) {
		u.PartSize = partMiBs * 1024 * 1024
	})
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(u.bucketName),
		Key:          aws.String(objectKey),
		Body:         contents,
		StorageClass: types.StorageClassDeepArchive,
	})
	if err != nil {
		return fmt.Errorf("couldn't upload large object to %v:%v: %w", u.bucketName, objectKey, err)
	}

	return err
}

func (u *uploader) pipeFolder(ctx context.Context, logger *slog.Logger, folderPath string) error {
	objectName := fmt.Sprintf("%s.zip", filepath.Base(folderPath))
	objectName = regexp.MustCompile(`\s+`).ReplaceAllString(objectName, "-")

	exists, err := u.checkObjectExists(ctx, logger, objectName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	pipeReader, pw := io.Pipe()

	errGroup, ctx := errgroup.WithContext(ctx)

	// upload
	errGroup.Go(func() error {
		logger.Info("uploading", "object", objectName)

		err := u.uploadLargeObject(ctx, objectName, pipeReader)
		if err != nil {
			return fmt.Errorf("put object: %w", err)
		}

		return pipeReader.Close()
	})

	// zip and write
	errGroup.Go(func() error {
		zipWriter := zip.NewWriter(pw)

		err := zipFolder(ctx, folderPath, zipWriter)
		if err != nil {
			return fmt.Errorf("compressing folder: %w", err)
		}
		logger.Info("finished zipping")
		if err := zipWriter.Close(); err != nil {
			return fmt.Errorf("closing zip file: %w", err)
		}
		if err := pw.Close(); err != nil {
			return err
		}
		logger.Info("writer closed")

		return nil
	})

	return errGroup.Wait()
}

func (u *uploader) checkObjectExists(ctx context.Context, logger *slog.Logger, objectName string) (bool, error) {
	_, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &u.bucketName,
		Key:    &objectName,
	})
	if err == nil {
		logger.Info("already exists")
		return true, nil
	}
	var nf *types.NotFound
	if !errors.As(err, &nf) {
		return false, fmt.Errorf("getting object attributes: %w", err)
	}
	return false, nil
}
