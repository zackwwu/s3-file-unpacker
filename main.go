package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/zackwwu/file-unpack-worker/internal"
	"github.com/zackwwu/file-unpack-worker/internal/config"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [FLAGS] [--source SOURCE_CLIP_URL | --bucket TARGET_BUCKET | --dir TARGET_DIRECTORY]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flagSourceURL := flag.String("source", "", "url of the S3 object")
	flagTargetBucket := flag.String("bucket", "", "bucket to unpack the object to")
	flagTargetDir := flag.String("dir", "", "directory in target bucket to unpack the object to")
	flagVerbose := flag.Bool("v", false, "enable debug and trace logs")
	flagBufferMemoryMB := flag.Int64("memory", 64, "the size of memory for keeping local copy of files, in MB")
	flagMaxParallelUpload := flag.Int64("p", 8, "the max number of parallel uploads, no more then 16")
	flag.Parse()

	if flag.NArg() == 0 && (*flagSourceURL == "" || *flagTargetBucket == "" || *flagTargetDir == "") {
		flag.Usage()
		os.Exit(2)
	}

	ctx := context.Background()
	log := zerolog.New(zerolog.NewConsoleWriter()).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	if *flagVerbose {
		log = log.Level(zerolog.TraceLevel)
	}

	ctx = log.WithContext(ctx)

	log.Info().
		Str("sourceURL", *flagSourceURL).
		Str("targetBucket", *flagTargetBucket).
		Str("targetDir", *flagTargetDir).
		Bool("verbose", *flagVerbose).
		Int64("bufferMemoryMB", *flagBufferMemoryMB).
		Int64("maxParallelUpload", *flagMaxParallelUpload).
		Msg("initializing")

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("unable to load config")
	}

	bufferMemoryBytes := (*flagBufferMemoryMB) << 20
	r := internal.Setup(&log, cfg, bufferMemoryBytes, *flagMaxParallelUpload)
	log.Info().Msg("running task")
	if err := r.Start(ctx, *flagSourceURL, *flagTargetBucket, *flagTargetDir); err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to run task")
	}
}
