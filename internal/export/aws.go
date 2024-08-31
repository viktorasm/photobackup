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

var _ Backend = (*s3destination)(nil)

func NewS3Destination(ctx context.Context, bucketName string) (Backend, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating s3 client: %w", err)
	}

	return &s3destination{
		bucketName: bucketName,
		client:     s3.NewFromConfig(awsConfig),
	}, nil
}

func (u *s3destination) Exists(ctx context.Context, objectName string) (bool, error) {
	_, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &u.bucketName,
		Key:    &objectName,
	})
	if err == nil {
		return true, nil
	}
	var nf *types.NotFound
	if !errors.As(err, &nf) {
		return false, fmt.Errorf("getting object attributes: %w", err)
	}
	return false, nil
}

func (u *s3destination) Write(ctx context.Context, objectKey string, contents io.Reader) error {
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
