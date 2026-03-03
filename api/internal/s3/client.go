package s3client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Client wraps the AWS S3 SDK client for file deletion operations.
type Client struct {
	svc    *s3.Client
	bucket string
}

// New creates a Client by loading AWS credentials from the environment.
func New(ctx context.Context, bucket string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("s3client.New: load config: %w", err)
	}

	return &Client{svc: s3.NewFromConfig(cfg), bucket: bucket}, nil
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
