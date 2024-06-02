package storage

import (
	"bytes"
	"fmt"
	"github.com/HyperGAI/serving-webhook/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
)

type Store interface {
	Upload(fileReader io.Reader, fileKey string) (string, error)
	PutObject(fileReader io.Reader, fileKey string) (string, error)
}

type S3Store struct {
	config   utils.Config
	uploader *s3manager.Uploader
	svc      *s3.S3
}

func NewS3Store(config utils.Config) (Store, error) {
	awsConfig := aws.Config{Region: aws.String(config.AWSRegion)}
	if config.AWSAccessKeyID != "" && config.AWSSecretAccessKey != "" {
		awsConfig = aws.Config{
			Region: aws.String(config.AWSRegion),
			Credentials: credentials.NewStaticCredentials(
				config.AWSAccessKeyID, config.AWSSecretAccessKey, ""),
		}
	}
	sess := session.Must(session.NewSession(&awsConfig))
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.Concurrency = 2
	})
	svc := s3.New(sess, aws.NewConfig().
		WithMaxRetries(3).
		WithS3UseAccelerate(config.AWSS3UseAccelerate),
	)
	return &S3Store{config: config, uploader: uploader, svc: svc}, nil
}

// Need to enable S3 ACL and set the block public access ACL permissions
// Use Gateway Endpoints for S3:
// Check https://docs.aws.amazon.com/vpc/latest/privatelink/vpc-endpoints-s3.html
// TODO: Bug, uploaded file becomes empty sometimes, https://github.com/aws/aws-sdk-go/issues/1962
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

func (uploader *S3Store) PutObject(fileReader io.Reader, fileKey string) (string, error) {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, fileReader); err != nil {
		return "", fmt.Errorf("failed to read file, %v", err)
	}

	data := buf.Bytes()
	if len(data) < 8 {
		return "", fmt.Errorf("file content length is less than 8, actual length: %d", len(data))
	}
	_, err := uploader.svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(uploader.config.AWSBucket),
		Key:    aws.String(fileKey),
		Body:   bytes.NewReader(data),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file, %v", err)
	}
	location := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		uploader.config.AWSBucket, uploader.config.AWSRegion, fileKey)
	return location, nil
}
