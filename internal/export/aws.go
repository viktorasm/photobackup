package export

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"io"
)

type s3destination struct {
	client     *s3.Client
	bucketName string
}

var _ Destination = (*s3destination)(nil)

func NewS3Destination(ctx context.Context, bucketName string) (Destination, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating s3 client: %w", err)
	}

	return &s3destination{
		bucketName: bucketName,
		client:     s3.NewFromConfig(awsConfig),
	}, nil
}

func (u *s3destination) Exists(ctx context.Context, path string, hash string) (bool, error) {
	resp, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &u.bucketName,
		Key:    &path,
	})
	if err != nil {
		var nf *types.NotFound
		if errors.As(err, &nf) {
			return false, nil
		}
		return false, fmt.Errorf("getting object attributes: %w", err)
	}

	metadataHash, hashExists := resp.Metadata["hash"]
	if !hashExists || metadataHash != hash {
		println("hash mismatch", hash, metadataHash)
		return true, nil // we probably don't want to rewrite with smaller file
	}
	return true, nil
}

func (u *s3destination) Write(ctx context.Context, name string, hash string, source io.Reader) error {
	var partMiBs int64 = 10
	uploader := manager.NewUploader(u.client, func(u *manager.Uploader) {
		u.PartSize = partMiBs * 1024 * 1024
	})
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(u.bucketName),
		Key:          aws.String(name),
		Body:         source,
		StorageClass: types.StorageClassDeepArchive,
		Metadata: map[string]string{
			"hash": hash,
		},
	})
	if err != nil {
		return fmt.Errorf("couldn't upload large object to %v:%v: %w", u.bucketName, name, err)
	}

	return err
}
