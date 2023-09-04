package storage

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/yangwenz/model-webhook/utils"
	"io"
)

type Store interface {
	Upload(fileReader io.Reader, fileKey string) (string, error)
}

type S3Store struct {
	config   utils.Config
	uploader *s3manager.Uploader
}

func NewS3Store(config utils.Config) (Store, error) {
	awsConfig := aws.Config{Region: aws.String(config.AWSRegion)}
	sess := session.Must(session.NewSession(&awsConfig))
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.Concurrency = 5
	})
	return &S3Store{config: config, uploader: uploader}, nil
}

// Need to enable S3 ACL and set the block public access ACL permissions
func (uploader *S3Store) Upload(fileReader io.Reader, fileKey string) (string, error) {
	result, err := uploader.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(uploader.config.AWSBucket),
		Key:    aws.String(fileKey),
		Body:   fileReader,
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}
	return result.Location, nil
}
