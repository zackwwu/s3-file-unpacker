package services

import (
	"context"
	"io"
)

type SourceStorage interface {
	Get(ctx context.Context, url string) (io.ReadCloser, error)
}
