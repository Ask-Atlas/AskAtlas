package s3client

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
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

	// Use path-style addressing ({endpoint}/{bucket}/{key}) so presigned URLs
	// stay on the endpoint hostname. Garage fronts behind a single-level
	// wildcard TLS cert that does not cover virtual-hosted {bucket}.{endpoint}.
	svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		// Strip SDK-added proxy-fragile headers RIGHT BEFORE the v4
		// signer runs. Our Garage instance is fronted by a CDN
		// reverse-proxy that normalises / rewrites a small set of
		// request headers on their way to the origin, so any header
		// the SDK signs but the proxy mutates invalidates the sig at
		// Garage. AWS CLI (botocore) dodges this by signing only
		// `host; x-amz-content-sha256; x-amz-date`; the Go SDK signs
		// everything it added, including `amz-sdk-invocation-id`,
		// `amz-sdk-request`, and `accept-encoding` — exactly the
		// headers the proxy rewrites. Deleting them before the signer
		// step matches the CLI's canonical request and unblocks every
		// direct S3 call (DeleteObject, GetObject, ListObjectsV2).
		//
		// Safety for retry tracking: the SDK carries invocation / retry
		// metadata in `context.Context` (not in wire-response headers),
		// so stripping the headers from the outbound request has no
		// effect on retry classification. Each attempt's finalize
		// chain runs in order {RetryMetricsHeader (re-sets
		// Amz-Sdk-Request) → this strip → Signing}, so every attempt's
		// signed canonical-request is clean.
		o.APIOptions = append(o.APIOptions, func(stack *middleware.Stack) error {
			return stack.Finalize.Insert(middleware.FinalizeMiddlewareFunc(
				"GarageStripProxyFragileHeaders",
				func(
					ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler,
				) (middleware.FinalizeOutput, middleware.Metadata, error) {
					if req, ok := in.Request.(*smithyhttp.Request); ok {
						req.Header.Del("Accept-Encoding")
						req.Header.Del("Amz-Sdk-Invocation-Id")
						req.Header.Del("Amz-Sdk-Request")
					}
					return next.HandleFinalize(ctx, in)
				},
			), "Signing", middleware.Before)
		})
	})
	return &Client{
		svc:       svc,
		presigner: s3.NewPresignClient(svc),
		bucket:    bucket,
	}, nil
}

// Bucket returns the bucket name the client was constructed against.
// Used by out-of-package callers that need to coordinate with the
// client (e.g., the seed-demo cleanup, which deletes by s3_key +
// bucket lookup).
func (c *Client) Bucket() string { return c.bucket }

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
