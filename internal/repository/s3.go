package repository

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	s3     *s3.Client
	tm     *transfermanager.Client
	bucket string
}

func NewS3Client(ctx context.Context, endpoint, accessKey, secretKey, bucket string) (*S3Client, error) {
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

	return &S3Client{
		s3:     s3Client,
		tm:     transfermanager.New(s3Client),
		bucket: bucket,
	}, nil
}

func (c *S3Client) Put(ctx context.Context, key string, r io.Reader) error {
	_, err := c.tm.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   r,
	})
	return err
}

func (c *S3Client) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil // вызывающий обязан закрыть
}

func (c *S3Client) Delete(ctx context.Context, key string) error {
	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	return err
}
