package s3client

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// presignExpiry is the validity duration for presigned upload URLs.
const presignExpiry = 15 * time.Minute

// Client wraps the AWS S3 SDK client for file operations.
type Client struct {
	svc       *s3.Client
	presigner *s3.PresignClient
	bucket    string
}

// New creates a Client by loading AWS credentials from the environment.
func New(ctx context.Context, bucket string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("s3client.New: load config: %w", err)
	}

	svc := s3.NewFromConfig(cfg)
	return &Client{
		svc:       svc,
		presigner: s3.NewPresignClient(svc),
		bucket:    bucket,
	}, nil
}

// GeneratePresignedPutURL creates a presigned S3 PUT URL for uploading a file.
func (c *Client) GeneratePresignedPutURL(ctx context.Context, key, contentType string, contentLength int64) (string, error) {
	req, err := c.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(contentLength),
	}, s3.WithPresignExpires(presignExpiry))
	if err != nil {
		return "", fmt.Errorf("s3client.GeneratePresignedPutURL: %w", err)
	}

	return req.URL, nil
}

// DeleteObject removes the object at the given S3 key.
func (c *Client) DeleteObject(ctx context.Context, key string) error {
	_, err := c.svc.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3client.DeleteObject: %w", err)
	}

	return nil
}
