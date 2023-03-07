package config

import (
	"context"
	"time"

	"github.com/Netflix/go-env"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/zackwwu/file-unpack-worker/internal/utils"
)

const ()

// TODO: use https://github.com/spf13/cobra to work with env vars, command arguments and flags
type envConfig struct {
	AWSClientTimeoutSec int `env:"AWS_CLIENT_TIMEOUT_SEC,default=30"`
	AWSClientMaxRetries int `env:"AWS_CLIENT_MAX_RETRIES,default=3"`
}

func (ec envConfig) Validate() error {
	return validation.ValidateStruct(&ec,
		validation.Field(&ec.AWSClientTimeoutSec, validation.Required, validation.Min(1)),
		validation.Field(&ec.AWSClientMaxRetries, validation.Required, validation.Min(1)),
	)
}

func loadEnvConfig() (*envConfig, error) {
	cfg := &envConfig{}
	_, err := env.UnmarshalFromEnviron(cfg)
	return cfg, err
}

type Config interface {
	AWSConfig() (aws.Config, error)
}

type config struct {
	// TODO: refactor, it should not be envConf, it's a config that read from env var and cli flags
	envConfig *envConfig
}

func NewConfig() (Config, error) {
	cfg, err := loadEnvConfig()
	if err != nil {
		return nil, err
	}

	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	return &config{envConfig: cfg}, nil
}

func (cfg *config) AWSConfig() (aws.Config, error) {
	var (
		clientTimeout = time.Duration(cfg.envConfig.AWSClientTimeoutSec) * time.Second
		maxRetries    = uint(cfg.envConfig.AWSClientMaxRetries)
	)

	return awsConfig.LoadDefaultConfig(context.Background(),
		awsConfig.WithHTTPClient(utils.NewAWSHTTPClient(
			clientTimeout,
			maxRetries),
		),
	)
}
