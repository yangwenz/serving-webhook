package utils

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
)

func GetAWSSession(region string, AWSAccessKeyID string, AWSSecretAccessKey string) *session.Session {
	awsConfig := aws.Config{Region: aws.String(region)}
	if AWSAccessKeyID != "" && AWSSecretAccessKey != "" {
		awsConfig = aws.Config{
			Region: aws.String(region),
			Credentials: credentials.NewStaticCredentials(
				AWSAccessKeyID, AWSSecretAccessKey, ""),
		}
	}
	return session.Must(session.NewSession(&awsConfig))
}

func GetS3Service(
	region string,
	AWSS3UseAccelerate bool,
	AWSAccessKeyID string,
	AWSSecretAccessKey string,
) *s3.S3 {
	sess := GetAWSSession(region, AWSAccessKeyID, AWSSecretAccessKey)
	awsConfig := &aws.Config{}
	awsConfig.WithS3UseAccelerate(AWSS3UseAccelerate)
	return s3.New(sess, awsConfig)
}

func UploadToS3(
	filepath string,
	bucket string,
	region string,
	key string,
	AWSAccessKeyID string,
	AWSSecretAccessKey string,
) error {
	sess := GetAWSSession(region, AWSAccessKeyID, AWSSecretAccessKey)
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.Concurrency = 16
	})

	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", filepath, err)
	}

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   f,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}
	fmt.Printf("file uploaded to, %s\n", result.Location)
	return nil
}
