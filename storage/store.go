package storage

import "io"

type Store interface {
	Upload(fileReader io.Reader, fileKey string) (string, error)
}
