package sourcestorage

import (
	"context"
	"io"

	"emperror.dev/errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/zackwwu/file-unpack-worker/internal/utils"
)

type Provider struct {
	client *s3.Client
}

func New(s3Client *s3.Client) *Provider {
	return &Provider{
		client: s3Client,
	}
}

func (p *Provider) Get(ctx context.Context, url string) (_ io.ReadCloser, err error) {
	bucket, key, err := utils.ParseS3URL(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse clip url")
	}

	pipeR, pipeW := io.Pipe()
	readCloseErr := make(chan error)

	readCloser := &sourceClipReadCloser{
		pipeReader: pipeR,
		errCh:      readCloseErr,
	}

	go func() {

		downloader := manager.NewDownloader(p.client)
		_, err = downloader.Download(ctx, &sourceClipWriteAt{pipeW}, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})

		pipeW.Close()
		readCloseErr <- err
	}()

	return readCloser, nil
}
