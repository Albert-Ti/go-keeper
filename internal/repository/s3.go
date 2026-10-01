package repository

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	client *transfermanager.Client
}

func NewS3Client(ctx context.Context, endpoint, accessKey, secretKey string) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	tmClient := transfermanager.New(s3Client) // ← оборачиваем s3.Client в transfermanager.Client

	return &S3Client{client: tmClient}, nil
}

func (s3c *S3Client) Put(ctx context.Context, key string, r io.Reader) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := s3c.client.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: aws.String("test"),
		Key:    aws.String(key),
		Body:   r,
	})
	return err
}

func (s3c *S3Client) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, nil
}

func (s3c *S3Client) Delete(ctx context.Context, key string) error {
	return nil
}
