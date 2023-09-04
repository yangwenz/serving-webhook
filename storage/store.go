package storage

type Store interface {
	Upload(fileBuffer []byte, fileKey string) (string, error)
}
