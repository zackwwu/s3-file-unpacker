package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zackwwu/file-unpack-worker/internal/utils"
)

func TestS3URLParsing(t *testing.T) {
	t.Run("can correctly parse S3 URI", func(t *testing.T) {
		s3URL := "s3://test-bucket/key/of/object"
		bucket, key, err := utils.ParseS3URL(s3URL)

		assert.NoError(t, err)
		assert.Equal(t, "test-bucket", bucket)
		assert.Equal(t, "/key/of/object", key)
	})
	t.Run("can correctly parse object url", func(t *testing.T) {
		s3URL := "https://test-bucket.s3.ap-southeast-2.amazonaws.com/key/of/object"
		bucket, key, err := utils.ParseS3URL(s3URL)

		assert.NoError(t, err)
		assert.Equal(t, "test-bucket", bucket)
		assert.Equal(t, "/key/of/object", key)
	})
	t.Run("returns error on malformed S3 URL", func(t *testing.T) {
		s3URL := "https://bucket/key/of/object"
		bucket, key, err := utils.ParseS3URL(s3URL)

		assert.Error(t, err)
		assert.Empty(t, bucket)
		assert.Empty(t, key)
	})
}
