package services

import (
	"context"
	"io"
)

type TargetStorage interface {
	Put(ctx context.Context, bucket string, directory string, fileName string, reader io.Reader) error
}
