package clipunpack

import (
	"bytes"
	"context"
	"io"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/kamilsk/retry/v5"
	"github.com/kamilsk/retry/v5/backoff"
	"github.com/kamilsk/retry/v5/strategy"
	"github.com/rs/zerolog/log"
	"github.com/zackwwu/file-unpack-worker/internal/services"
	"github.com/zackwwu/file-unpack-worker/internal/utils/tarstream"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	bufferMemoryBytes int64
	maxParallelUpload int64
	sourceStorage     services.SourceStorage
	targetStorage     services.TargetStorage
}

func New(
	bufferMemoryBytes int64,
	maxParallelUpload int64,
	sourceStorage services.SourceStorage,
	targetStorage services.TargetStorage,
) *Service {
	return &Service{
		bufferMemoryBytes: bufferMemoryBytes,
		maxParallelUpload: maxParallelUpload,
		sourceStorage:     sourceStorage,
		targetStorage:     targetStorage,
	}
}

func (s *Service) RunTask(ctx context.Context, sourceURL string, targetBucket string, targetDirectory string) error {
	logger := log.Ctx(ctx)

	logger.Trace().Msg("creating reader from source file")

	readCloser, err := s.sourceStorage.Get(ctx, sourceURL)
	if err != nil {
		return errors.Wrap(err, "failed to create reader from source file")
	}

	// TODO: improve error handling
	defer readCloser.Close()

	remainingMemory := s.bufferMemoryBytes
	largeFileUploading := false
	unpacker := tarstream.New(readCloser, logger)

	var m sync.Mutex
	cond := sync.NewCond(&m)

	eg, egCtx := errgroup.WithContext(ctx)
	egCtx, egCancel := context.WithCancel(egCtx)
	eg.SetLimit(int(s.maxParallelUpload))

	for {
		out, err := unpacker.Next()
		if err != nil {
			egCancel()
			// TODO: revert
			break
		}

		egLogger := logger.With().
			Str("fileName", out.Name).
			Int64("size", out.Size).
			Logger()
		gCtx := log.Logger.WithContext(egCtx)

		{
			cond.L.Lock()
			for largeFileUploading || out.Size > remainingMemory {
				cond.Wait()
			}

			if out.Size > s.bufferMemoryBytes {
				largeFileUploading = true
			} else {
				remainingMemory -= out.Size
			}
			cond.L.Unlock()
		}

		eg.Go(func() error {
			defer func() {
				cond.L.Lock()
				if out.Size > s.bufferMemoryBytes {
					largeFileUploading = false
				} else {
					remainingMemory += out.Size
				}
				cond.L.Unlock()
				cond.Signal()
			}()

			if out.Size > s.bufferMemoryBytes {
				egLogger.Trace().Msg("uploading large file")
				err := s.uploadLargeFile(gCtx, targetBucket, targetDirectory, out.Name, out.Reader)
				return errors.Wrap(err, "failed to upload large file")
			}

			egLogger.Trace().Msg("uploading file with buffer")
			err := s.uploadFile(gCtx, targetBucket, targetDirectory, out.Name, out.Reader)
			return errors.Wrap(err, "failed to upload file")
		})
	}

	// TODO: clean up files
	return eg.Wait()
}

func (s *Service) uploadLargeFile(
	ctx context.Context,
	targetBucket string,
	targetDirectory string,
	name string,
	reader io.Reader,
) error {
	return s.targetStorage.Put(ctx, targetBucket, targetDirectory, name, reader)
}

func (s *Service) uploadFile(
	ctx context.Context,
	targetBucket string,
	targetDirectory string,
	name string,
	reader io.Reader,
) error {
	content, err := io.ReadAll(reader)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}

	readSeeker := bytes.NewReader(content)

	action := func(aCtx context.Context) (aErr error) {
		_, aErr = readSeeker.Seek(0, io.SeekStart)
		if aErr != nil {
			return errors.Wrap(aErr, "failed to seek file content")
		}

		aErr = s.targetStorage.Put(aCtx, targetBucket, targetDirectory, name, readSeeker)
		return errors.Wrap(aErr, "failed to put file")
	}

	return retry.Do(ctx, action,
		strategy.Limit(3),
		strategy.Backoff(
			backoff.BinaryExponential(500*time.Millisecond),
		),
	)
}
