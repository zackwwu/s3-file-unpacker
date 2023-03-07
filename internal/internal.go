package internal

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/zackwwu/file-unpack-worker/internal/config"
	"github.com/zackwwu/file-unpack-worker/internal/providers/sourcestorage"
	"github.com/zackwwu/file-unpack-worker/internal/providers/targetstorage"
	"github.com/zackwwu/file-unpack-worker/internal/services/clipunpack"

	"github.com/rs/zerolog"
)

type TaskRunner interface {
	RunTask(ctx context.Context, sourceURL string, targetBucket string, targetDirectory string) error
}

type Runner struct {
	runner TaskRunner
}

func (r *Runner) Start(ctx context.Context, source string, targetBucket string, targetDirectory string) error {
	return r.runner.RunTask(ctx, source, targetBucket, targetDirectory)
}

func Setup(
	logger *zerolog.Logger,
	cfg config.Config,
	bufferMemoryBytes int64,
	maxParallelUpload int64,
) *Runner {
	logger.Info().
		Msg("setting up task runner")

	awsCfg, err := cfg.AWSConfig()

	if err != nil {
		logger.Fatal().
			Err(err).
			Msg("unable to get AWS config")
	}

	s3Client := s3.NewFromConfig(awsCfg)
	sourceStorage := sourcestorage.New(s3Client)
	targetStorage := targetstorage.New(s3Client)

	runner := clipunpack.New(
		bufferMemoryBytes,
		maxParallelUpload,
		sourceStorage,
		targetStorage,
	)

	return &Runner{
		runner: runner,
	}
}
