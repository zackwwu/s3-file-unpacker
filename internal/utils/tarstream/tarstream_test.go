package tarstream_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zackwwu/file-unpack-worker/internal/utils/tarstream"
)

// TODO: improve the unit test by creating tar file at the beginning of the test
func TestTarStream(t *testing.T) {
	t.Run("can iterate through files", func(t *testing.T) {
		f, err := os.Open("../../../test/test.tar")
		require.NoError(t, err)
		require.NotNil(t, f)

		defer f.Close()

		logger := zerolog.Nop()
		unpacker := tarstream.New(f, &logger)
		for {
			out, err := unpacker.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("error: %v\n", err)
				break
			}

			fmt.Printf("file name: %s\n", out.Name)
			bytes, err := io.ReadAll(out.Reader)
			require.NoError(t, err)
			fmt.Printf("content: %s\n", string(bytes))
		}

		assert.Equal(t, 1, 2)
	})
}
