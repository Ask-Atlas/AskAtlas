package s3client_test

import (
	"context"
	"net/url"
	"testing"

	s3client "github.com/Ask-Atlas/AskAtlas/api/internal/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Presigning is fully client-side; stub creds so the SDK config chain resolves.
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
				"/askatlas-dev/uploads/abc/file.pdf",
				"X-Amz-Signature=",
				"X-Amz-Expires=900",
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
			require.NoError(t, err)
			assert.NotEmpty(t, parsed.Host)

			for _, want := range tc.wantFragments {
				assert.Contains(t, got, want)
			}
		})
	}
}

func TestGeneratePresignedGetURL_EmptyKey(t *testing.T) {
	withStubCreds(t)
	c, err := s3client.New(context.Background(), "askatlas-dev")
	require.NoError(t, err)

	_, err = c.GeneratePresignedGetURL(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "s3client.GeneratePresignedGetURL")
}
