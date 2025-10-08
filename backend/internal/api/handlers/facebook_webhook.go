// path: backend/internal/api/handlers/facebook_webhook.go
package handlers

import (
	"net/http"

	"github.com/techappsUT/social-queue/internal/social/adapters"
)

type FacebookWebhookHandlerHTTP struct {
	webhookHandler *adapters.FacebookWebhookHandler
}

func NewFacebookWebhookHandlerHTTP(webhookHandler *adapters.FacebookWebhookHandler) *FacebookWebhookHandlerHTTP {
	return &FacebookWebhookHandlerHTTP{
		webhookHandler: webhookHandler,
	}
}

// GET /api/webhooks/facebook - Webhook verification
func (h *FacebookWebhookHandlerHTTP) VerifyWebhook(w http.ResponseWriter, r *http.Request) {
	h.webhookHandler.VerifyWebhook(w, r)
}

// POST /api/webhooks/facebook - Webhook events
func (h *FacebookWebhookHandlerHTTP) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	h.webhookHandler.HandleWebhook(w, r)
}
