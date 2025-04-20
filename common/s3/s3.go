package s3

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

type S3Client struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	BucketName      string
}

type IS3Client interface {
	UploadFile(context.Context, string, []byte) (string, error)
}

func NewS3Client(accessKeyID, secretAccessKey, region, bucketName string) IS3Client {
	return &S3Client{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		Region:          region,
		BucketName:      bucketName,
	}
}

func (s *S3Client) createClient() (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(s.Region),
		Credentials: credentials.NewStaticCredentials(s.AccessKeyID, s.SecretAccessKey, ""),
	})
	if err != nil {
		logrus.Errorf("Failed to create AWS session: %v", err)
		return nil, err
	}

	return s3.New(sess), nil
}

func (s *S3Client) UploadFile(ctx context.Context, fileName string, data []byte) (string, error) {
	var (
		contentType      = "application/octet-stream"
		timeoutInSeconds = 60
	)

	client, err := s.createClient()
	if err != nil {
		logrus.Errorf("Failed to create S3 client: %v", err)
		return "", err
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutInSeconds)*time.Second)
	defer cancel()

	// Prepare the upload parameters
	params := &s3.PutObjectInput{
		Bucket:        aws.String(s.BucketName),
		Key:           aws.String(fileName),
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(int64(len(data))),
		ContentType:   aws.String(contentType),
	}

	// Upload the file
	_, err = client.PutObjectWithContext(ctx, params)
	if err != nil {
		logrus.Errorf("Failed to upload file to S3: %v", err)
		return "", err
	}

	// Generate the URL to the object
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.BucketName, s.Region, fileName)
	return url, nil
}
