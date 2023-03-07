package sourcestorage

import "io"

type sourceClipReadCloser struct {
	pipeReader *io.PipeReader
	errCh      chan error
}

func (rc *sourceClipReadCloser) Read(p []byte) (n int, err error) {
	return rc.pipeReader.Read(p)
}

func (rc *sourceClipReadCloser) Close() error {
	if err := rc.pipeReader.Close(); err != nil {
		return err
	}

	return <-rc.errCh
}

type sourceClipWriteAt struct {
	pipeWriter *io.PipeWriter
}

func (wa *sourceClipWriteAt) WriteAt(p []byte, offset int64) (n int, err error) {
	return wa.pipeWriter.Write(p)
}
