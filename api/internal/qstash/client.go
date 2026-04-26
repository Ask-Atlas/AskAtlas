package qstashclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/upstash/qstash-go"
)

// DeleteFileMessage is the payload published to the delete-file job queue.
type DeleteFileMessage struct {
	FileID      string `json:"file_id"`
	S3Key       string `json:"s3_key"`
	UserID      string `json:"user_id"`
	RequestedAt string `json:"requested_at"`
	Environment string `json:"environment,omitempty"`
}

// ExtractFileMessage is the payload published to the extract-file job
// queue (ASK-220). The worker re-fetches s3_key + mime_type from the DB
// to avoid trusting QStash redelivery to carry stale field values, but
// they're still in the body for log correlation when the DB row vanishes.
type ExtractFileMessage struct {
	FileID      string `json:"file_id"`
	S3Key       string `json:"s3_key"`
	MimeType    string `json:"mime_type"`
	UserID      string `json:"user_id"`
	RequestedAt string `json:"requested_at"`
	Environment string `json:"environment,omitempty"`
}

// Client wraps the QStash SDK for publishing job messages.
type Client struct {
	client     *qstash.Client
	jobBaseURL string
	env        string
}

// New creates a Client using the provided QStash token and base URL for job endpoints.
func New(token, jobBaseURL, env string) *Client {
	if env == "" {
		env = "unknown"
	}

	return &Client{
		client:     qstash.NewClient(token),
		jobBaseURL: jobBaseURL,
		env:        env,
	}
}

// PublishDeleteFile sends a delete-file message to QStash and returns the QStash message ID.
func (c *Client) PublishDeleteFile(ctx context.Context, msg DeleteFileMessage) (string, error) {
	if msg.Environment == "" {
		msg.Environment = c.env
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("qstashclient.PublishDeleteFile: marshal: %w", err)
	}

	res, err := c.client.Publish(qstash.PublishOptions{
		Url:             c.jobBaseURL + "/delete-file",
		Body:            string(body),
		Method:          http.MethodPost,
		ContentType:     "application/json",
		FailureCallback: c.jobBaseURL + "/delete-file-failed",
	})
	if err != nil {
		return "", fmt.Errorf("qstashclient.PublishDeleteFile: publish: %w", err)
	}

	return res.MessageId, nil
}

// PublishExtractFile sends an extract-file message to QStash (ASK-220).
// Mirrors PublishDeleteFile in shape so the file service can swap in
// either through the same QStashPublisher interface.
func (c *Client) PublishExtractFile(ctx context.Context, msg ExtractFileMessage) (string, error) {
	if msg.Environment == "" {
		msg.Environment = c.env
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("qstashclient.PublishExtractFile: marshal: %w", err)
	}

	res, err := c.client.Publish(qstash.PublishOptions{
		Url:             c.jobBaseURL + "/extract-file",
		Body:            string(body),
		Method:          http.MethodPost,
		ContentType:     "application/json",
		FailureCallback: c.jobBaseURL + "/extract-file-failed",
	})
	if err != nil {
		return "", fmt.Errorf("qstashclient.PublishExtractFile: publish: %w", err)
	}

	return res.MessageId, nil
}
