// path: backend/internal/social/adapters/facebook_webhook.go
package adapters

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// FacebookWebhookHandler handles Facebook webhook events
// Required for:
// - Real-time updates on page posts
// - Instagram mentions and comments
// - Page changes (e.g., page access revoked)
type FacebookWebhookHandler struct {
	appSecret   string
	verifyToken string // Set in Facebook App Dashboard
}

// NewFacebookWebhookHandler creates a new webhook handler
func NewFacebookWebhookHandler(appSecret, verifyToken string) *FacebookWebhookHandler {
	return &FacebookWebhookHandler{
		appSecret:   appSecret,
		verifyToken: verifyToken,
	}
}

// FacebookWebhookEvent represents a Facebook webhook payload
type FacebookWebhookEvent struct {
	Object string `json:"object"` // "page", "instagram", "user"
	Entry  []struct {
		ID        string                   `json:"id"`
		Time      int64                    `json:"time"`
		Changes   []FacebookWebhookChange  `json:"changes"`
		Messaging []FacebookMessagingEvent `json:"messaging"` // For Messenger events
	} `json:"entry"`
}

// FacebookWebhookChange represents a change event
type FacebookWebhookChange struct {
	Field string                 `json:"field"` // e.g., "feed", "comments", "ratings"
	Value map[string]interface{} `json:"value"`
}

// FacebookMessagingEvent represents Messenger events
type FacebookMessagingEvent struct {
	Sender    map[string]string      `json:"sender"`
	Recipient map[string]string      `json:"recipient"`
	Timestamp int64                  `json:"timestamp"`
	Message   map[string]interface{} `json:"message"`
}

// VerifyWebhook handles GET request for webhook verification
// Facebook sends this when you set up the webhook in App Dashboard
func (h *FacebookWebhookHandler) VerifyWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	if mode == "subscribe" && token == h.verifyToken {
		// Verification successful
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}

	http.Error(w, "Forbidden", http.StatusForbidden)
}

// HandleWebhook processes POST requests with webhook events
func (h *FacebookWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		http.Error(w, "Missing signature", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	if !h.verifySignature(body, signature) {
		http.Error(w, "Invalid signature", http.StatusForbidden)
		return
	}

	// Parse webhook event
	var event FacebookWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Process the event
	h.processWebhookEvent(&event)

	// Always respond with 200 OK quickly
	// Facebook expects response within 20 seconds
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("EVENT_RECEIVED"))
}

// verifySignature validates the X-Hub-Signature-256 header
func (h *FacebookWebhookHandler) verifySignature(body []byte, signatureHeader string) bool {
	// Remove "sha256=" prefix
	expectedSignature := signatureHeader[7:]

	// Compute HMAC
	mac := hmac.New(sha256.New, []byte(h.appSecret))
	mac.Write(body)
	actualSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSignature), []byte(actualSignature))
}

// processWebhookEvent handles different webhook event types
func (h *FacebookWebhookHandler) processWebhookEvent(event *FacebookWebhookEvent) {
	switch event.Object {
	case "page":
		h.processPageEvent(event)
	case "instagram":
		h.processInstagramEvent(event)
	case "user":
		h.processUserEvent(event)
	default:
		fmt.Printf("Unknown webhook object type: %s\n", event.Object)
	}
}

// processPageEvent handles Facebook Page events
func (h *FacebookWebhookHandler) processPageEvent(event *FacebookWebhookEvent) {
	for _, entry := range event.Entry {
		pageID := entry.ID

		for _, change := range entry.Changes {
			switch change.Field {
			case "feed":
				// New post, post edited, post deleted
				h.handlePageFeedChange(pageID, change.Value)
			case "comments":
				// New comment on page post
				h.handlePageCommentChange(pageID, change.Value)
			case "ratings":
				// New page rating/review
				h.handlePageRatingChange(pageID, change.Value)
			case "live_videos":
				// Live video status change
				h.handleLiveVideoChange(pageID, change.Value)
			}
		}
	}
}

// processInstagramEvent handles Instagram Business account events
func (h *FacebookWebhookHandler) processInstagramEvent(event *FacebookWebhookEvent) {
	for _, entry := range event.Entry {
		igAccountID := entry.ID

		for _, change := range entry.Changes {
			switch change.Field {
			case "comments":
				// New comment on Instagram post
				h.handleInstagramCommentChange(igAccountID, change.Value)
			case "mentions":
				// User mentioned in Instagram caption/comment
				h.handleInstagramMentionChange(igAccountID, change.Value)
			case "story_insights":
				// Instagram story insights available
				h.handleInstagramStoryInsights(igAccountID, change.Value)
			}
		}
	}
}

// processUserEvent handles user-level events
func (h *FacebookWebhookHandler) processUserEvent(event *FacebookWebhookEvent) {
	// Handle user permission changes, etc.
	fmt.Printf("User event received: %+v\n", event)
}

// Event handler implementations
func (h *FacebookWebhookHandler) handlePageFeedChange(pageID string, value map[string]interface{}) {
	// TODO: Implement based on your needs
	// Example: Update post status in database, sync analytics
	fmt.Printf("Page %s feed change: %+v\n", pageID, value)
}

func (h *FacebookWebhookHandler) handlePageCommentChange(pageID string, value map[string]interface{}) {
	// TODO: Store comment in database, notify user, etc.
	fmt.Printf("Page %s comment change: %+v\n", pageID, value)
}

func (h *FacebookWebhookHandler) handlePageRatingChange(pageID string, value map[string]interface{}) {
	fmt.Printf("Page %s rating change: %+v\n", pageID, value)
}

func (h *FacebookWebhookHandler) handleLiveVideoChange(pageID string, value map[string]interface{}) {
	fmt.Printf("Page %s live video change: %+v\n", pageID, value)
}

func (h *FacebookWebhookHandler) handleInstagramCommentChange(igAccountID string, value map[string]interface{}) {
	fmt.Printf("Instagram %s comment change: %+v\n", igAccountID, value)
}

func (h *FacebookWebhookHandler) handleInstagramMentionChange(igAccountID string, value map[string]interface{}) {
	fmt.Printf("Instagram %s mention: %+v\n", igAccountID, value)
}

func (h *FacebookWebhookHandler) handleInstagramStoryInsights(igAccountID string, value map[string]interface{}) {
	fmt.Printf("Instagram %s story insights: %+v\n", igAccountID, value)
}
