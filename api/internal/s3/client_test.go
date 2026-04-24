package s3client_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	s3client "github.com/Ask-Atlas/AskAtlas/api/internal/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Presigning happens entirely client-side -- no network, no creds
// validation. We inject dummy creds via env so the AWS config chain
// resolves locally and the signer has something to work with.
func withStubCreds(t *testing.T) {
	t.Helper()
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIATEST")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "secret-test") // pragma: allowlist secret
	t.Setenv("AWS_REGION", "us-east-1")
}

func TestGeneratePresignedGetURL(t *testing.T) {
	withStubCreds(t)

	tests := []struct {
		name   string
		bucket string
		key    string
		// wantFragments are substrings that MUST appear in the final URL.
		// Avoids binding the test to the exact parameter order emitted by
		// the AWS SDK signer.
		wantFragments []string
	}{
		{
			name:   "simple key",
			bucket: "askatlas-dev",
			key:    "uploads/abc/file.pdf",
			wantFragments: []string{
				"/askatlas-dev/uploads/abc/file.pdf", // path-style addressing
				"X-Amz-Signature=",
				"X-Amz-Expires=900", // presignExpiry = 15 min
				"X-Amz-Algorithm=AWS4-HMAC-SHA256",
			},
		},
		{
			name:   "nested key with spaces gets url-encoded",
			bucket: "askatlas-dev",
			key:    "uploads/abc/a b c.pdf",
			wantFragments: []string{
				"/askatlas-dev/uploads/abc/a%20b%20c.pdf",
				"X-Amz-Signature=",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := s3client.New(context.Background(), tc.bucket)
			require.NoError(t, err)

			got, err := c.GeneratePresignedGetURL(context.Background(), tc.key)
			require.NoError(t, err)

			parsed, err := url.Parse(got)
			require.NoError(t, err, "presigned URL must be parseable")
			assert.NotEmpty(t, parsed.Host, "presigned URL must have a host")

			for _, want := range tc.wantFragments {
				assert.True(t,
					strings.Contains(got, want),
					"expected URL to contain %q, got %q", want, got,
				)
			}
		})
	}
}

// Empty key: AWS SDK validates and returns an error at presign time,
// so our wrapper should surface it wrapped.
func TestGeneratePresignedGetURL_EmptyKey(t *testing.T) {
	withStubCreds(t)
	c, err := s3client.New(context.Background(), "askatlas-dev")
	require.NoError(t, err)

	_, err = c.GeneratePresignedGetURL(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "s3client.GeneratePresignedGetURL")
}
