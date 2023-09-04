package storage

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/yangwenz/model-webhook/utils"
)

type Uploader struct {
	config   utils.Config
	uploader *s3manager.Uploader
}

func NewUploader(config utils.Config) (*Uploader, error) {
	awsConfig := aws.Config{Region: aws.String(config.AWSRegion)}
	sess := session.Must(session.NewSession(&awsConfig))
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.Concurrency = 5
	})
	return &Uploader{config: config, uploader: uploader}, nil
}

func (uploader *Uploader) Upload(fileBuffer []byte, fileKey string) (string, error) {
	result, err := uploader.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(uploader.config.AWSBucket),
		Key:    aws.String(fileKey),
		Body:   bytes.NewReader(fileBuffer),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}
	return result.Location, nil
}
