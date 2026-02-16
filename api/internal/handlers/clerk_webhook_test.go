package handlers_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/internal/clerk"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	mock_handlers "github.com/Ask-Atlas/AskAtlas/api/internal/handlers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClerkHandler_Webhook_ValidUserCreatedEvent(t *testing.T) {
	t.Parallel()
	body := `{"type": "user.created", "data": {"id": "user_123"}}`

	mockService := mock_handlers.NewMockClerkService(t)
	mockService.EXPECT().
		HandleWebhookEvent(mock.Anything, mock.MatchedBy(func(e clerk.Event) bool {
			return e.GetType() == clerk.UserCreated
		})).
		Return(nil)

	h := handlers.NewClerkWebhookHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/clerk", bytes.NewBufferString(body))
	req.Header.Set("svix-id", "msg_123")
	w := httptest.NewRecorder()

	h.Webhook(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestClerkHandler_Webhook_InvalidJson(t *testing.T) {
	t.Parallel()
	body := `{invalid-json}`

	mockService := mock_handlers.NewMockClerkService(t)

	h := handlers.NewClerkWebhookHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/clerk", bytes.NewBufferString(body))
	req.Header.Set("svix-id", "msg_123")
	w := httptest.NewRecorder()

	h.Webhook(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestClerkHandler_Webhook_ServiceError(t *testing.T) {
	t.Parallel()
	body := `{"type": "user.created", "data": {"id": "user_123"}}`

	mockService := mock_handlers.NewMockClerkService(t)
	mockService.EXPECT().
		HandleWebhookEvent(mock.Anything, mock.Anything).
		Return(errors.New("handler error"))

	h := handlers.NewClerkWebhookHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/clerk", bytes.NewBufferString(body))
	req.Header.Set("svix-id", "msg_123")
	w := httptest.NewRecorder()

	h.Webhook(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestClerkHandler_Webhook_UserNotFoundOnDelete(t *testing.T) {
	t.Parallel()
	body := `{"type": "user.deleted", "data": {"id": "user_missing", "deleted": true}}`

	mockService := mock_handlers.NewMockClerkService(t)
	mockService.EXPECT().
		HandleWebhookEvent(mock.Anything, mock.Anything).
		Return(clerk.ErrUserNotFound)

	h := handlers.NewClerkWebhookHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/clerk", bytes.NewBufferString(body))
	req.Header.Set("svix-id", "msg_123")
	w := httptest.NewRecorder()

	h.Webhook(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
