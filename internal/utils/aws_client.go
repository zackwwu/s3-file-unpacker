package utils

import (
	"net/http"
	"time"

	httpc "github.com/zackwwu/http-client-go"
)

type AWSHTTPClient struct {
	client *httpc.Client
}

func (c *AWSHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

func NewAWSHTTPClient(reqTimeout time.Duration, maxRetries uint) *AWSHTTPClient {
	return &AWSHTTPClient{
		client: httpc.New(
			httpc.WithStandardRetryPolicy(reqTimeout, maxRetries),
		),
	}
}
