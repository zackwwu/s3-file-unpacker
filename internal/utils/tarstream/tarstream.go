package tarstream

import (
	"archive/tar"
	"io"
	"path/filepath"
	"strings"

	"emperror.dev/errors"

	"github.com/rs/zerolog"
)

type Unpacker struct {
	logger *zerolog.Logger
	reader *tar.Reader
}

func New(tarStream io.Reader, logger *zerolog.Logger) *Unpacker {
	return &Unpacker{
		reader: tar.NewReader(tarStream),
		logger: logger,
	}
}

type UnpackerOut struct {
	Name   string
	Size   int64
	Reader io.Reader
}

func (t *Unpacker) Next() (UnpackerOut, error) {
	header, err := t.reader.Next()
	if err != nil {
		return UnpackerOut{}, err
	}

	logger := t.logger.With().
		Str("typeflag", "dir").
		Str("headerName", header.Name).
		Logger()

	switch header.Typeflag {
	case tar.TypeDir:
		logger.Trace().
			Msg("directory, ignore")
		return UnpackerOut{}, nil
	case tar.TypeReg:
		if strings.HasPrefix(filepath.Base(header.Name), "._") {
			logger.Trace().Msg("auxiliary information file, ignore")
			return UnpackerOut{}, nil
		}
		logger.Trace().Msg("valid regular file, return for uploading")
		return UnpackerOut{
			Name:   header.Name,
			Size:   header.Size,
			Reader: t.reader,
		}, nil

	default:
		return UnpackerOut{}, errors.Errorf("unknown type flag for tar: %b", header.Typeflag)
	}
}
