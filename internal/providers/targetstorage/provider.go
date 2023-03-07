package targetstorage

import (
	"context"
	"io"
	"path/filepath"

	"emperror.dev/errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

type Provider struct {
	client *s3.Client
}

func New(s3Client *s3.Client) *Provider {
	return &Provider{
		client: s3Client,
	}
}

func (p *Provider) Put(
	ctx context.Context,
	bucket string,
	directory string,
	fileName string,
	reader io.Reader,
) error {
	key := filepath.Join(directory, fileName)

	uploader := manager.NewUploader(p.client)

	logger := log.Ctx(ctx).With().Str("key", key).Logger()
	logger.Trace().Msg("uploading file")

	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   reader,
	})

	return errors.Wrap(err, "failed to upload file")
}
