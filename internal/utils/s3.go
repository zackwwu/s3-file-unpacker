package utils

import (
	"net/url"
	"strings"

	"emperror.dev/errors"
)

// expecting s3://{bucket-name}/{key} or https://{bucket-name}.s3.{region}/{key}
func ParseS3URL(s3URL string) (bucket string, key string, err error) {
	u, err := url.Parse(s3URL)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to parse s3 url")
	}

	switch u.Scheme {
	case "s3":
		return parseS3URI(u)
	case "https":
		return parseObjectURL(u)
	default:
		return "", "", errors.New("malformed s3 url")
	}
}

func parseS3URI(u *url.URL) (bucket string, key string, err error) {
	return u.Host, u.Path, nil
}

func parseObjectURL(u *url.URL) (bucket string, key string, err error) {
	segments := strings.Split(u.Host, ".s3")
	if len(segments) <= 1 {
		return "", "", errors.New("malformed s3 url")
	}

	return segments[0], u.Path, nil
}
