// ============================================================================
// FILE 2: backend/internal/handlers/post_handler.go
// ============================================================================
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/techappsUT/social-queue/internal/application/post"
	postDomain "github.com/techappsUT/social-queue/internal/domain/post"
	"github.com/techappsUT/social-queue/internal/middleware"
)

type PostHandler struct {
	createDraftUC  *post.CreateDraftUseCase
	schedulePostUC *post.SchedulePostUseCase
	updatePostUC   *post.UpdatePostUseCase
	deletePostUC   *post.DeletePostUseCase
	getPostUC      *post.GetPostUseCase
	listPostsUC    *post.ListPostsUseCase
	publishNowUC   *post.PublishNowUseCase
}

func NewPostHandler(
	createDraftUC *post.CreateDraftUseCase,
	schedulePostUC *post.SchedulePostUseCase,
	updatePostUC *post.UpdatePostUseCase,
	deletePostUC *post.DeletePostUseCase,
	getPostUC *post.GetPostUseCase,
	listPostsUC *post.ListPostsUseCase,
	publishNowUC *post.PublishNowUseCase,
) *PostHandler {
	return &PostHandler{
		createDraftUC:  createDraftUC,
		schedulePostUC: schedulePostUC,
		updatePostUC:   updatePostUC,
		deletePostUC:   deletePostUC,
		getPostUC:      getPostUC,
		listPostsUC:    listPostsUC,
		publishNowUC:   publishNowUC,
	}
}

// ============================================================================
// POST /api/v2/posts - Create Draft
// ============================================================================

func (h *PostHandler) CreateDraft(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input post.CreateDraftInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input.AuthorID = userID

	output, err := h.createDraftUC.Execute(r.Context(), input)
	if err != nil {
		switch err {
		case postDomain.ErrEmptyContent:
			respondError(w, http.StatusBadRequest, "content cannot be empty")
		case postDomain.ErrNoPlatformsSelected:
			respondError(w, http.StatusBadRequest, "at least one platform must be selected")
		case postDomain.ErrInvalidPlatform:
			respondError(w, http.StatusBadRequest, "invalid platform selected")
		default:
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	respondCreated(w, output)
}

// ============================================================================
// GET /api/v2/posts/:id - Get Post
// ============================================================================

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	postIDStr := chi.URLParam(r, "id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid post ID")
		return
	}

	input := post.GetPostInput{
		PostID: postID,
		UserID: userID,
	}

	output, err := h.getPostUC.Execute(r.Context(), input)
	if err != nil {
		if err == postDomain.ErrPostNotFound {
			respondError(w, http.StatusNotFound, "post not found")
		} else {
			respondError(w, http.StatusForbidden, err.Error())
		}
		return
	}

	respondSuccess(w, output)
}

// ============================================================================
// PUT /api/v2/posts/:id - Update Post
// ============================================================================

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	postIDStr := chi.URLParam(r, "id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid post ID")
		return
	}

	var input post.UpdatePostInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input.PostID = postID
	input.UserID = userID

	output, err := h.updatePostUC.Execute(r.Context(), input)
	if err != nil {
		switch err {
		case postDomain.ErrPostNotFound:
			respondError(w, http.StatusNotFound, "post not found")
		case postDomain.ErrCannotEditPublished:
			respondError(w, http.StatusBadRequest, "cannot edit published post")
		default:
			respondError(w, http.StatusForbidden, err.Error())
		}
		return
	}

	respondSuccess(w, output)
}

// ============================================================================
// DELETE /api/v2/posts/:id - Delete Post
// ============================================================================

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	postIDStr := chi.URLParam(r, "id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid post ID")
		return
	}

	input := post.DeletePostInput{
		PostID: postID,
		UserID: userID,
	}

	if err := h.deletePostUC.Execute(r.Context(), input); err != nil {
		if err == postDomain.ErrPostNotFound {
			respondError(w, http.StatusNotFound, "post not found")
		} else {
			respondError(w, http.StatusForbidden, err.Error())
		}
		return
	}

	respondNoContent(w)
}

// ============================================================================
// POST /api/v2/posts/:id/schedule - Schedule Post
// ============================================================================

func (h *PostHandler) SchedulePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	postIDStr := chi.URLParam(r, "id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid post ID")
		return
	}

	var input post.SchedulePostInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input.PostID = postID
	input.UserID = userID

	output, err := h.schedulePostUC.Execute(r.Context(), input)
	if err != nil {
		switch err {
		case postDomain.ErrPostNotFound:
			respondError(w, http.StatusNotFound, "post not found")
		case postDomain.ErrScheduleTimeInPast:
			respondError(w, http.StatusBadRequest, "schedule time cannot be in the past")
		default:
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	respondSuccess(w, output)
}

// ============================================================================
// POST /api/v2/posts/:id/publish - Publish Now
// ============================================================================

func (h *PostHandler) PublishNow(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	postIDStr := chi.URLParam(r, "id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid post ID")
		return
	}

	input := post.PublishNowInput{
		PostID: postID,
		UserID: userID,
	}

	output, err := h.publishNowUC.Execute(r.Context(), input)
	if err != nil {
		if err == postDomain.ErrPostNotFound {
			respondError(w, http.StatusNotFound, "post not found")
		} else {
			respondError(w, http.StatusForbidden, err.Error())
		}
		return
	}

	respondSuccess(w, output)
}

// ============================================================================
// GET /api/v2/teams/:teamId/posts - List Posts
// ============================================================================

func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	teamIDStr := chi.URLParam(r, "teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid team ID")
		return
	}

	// Parse query params
	var offset, limit int
	fmt.Sscanf(r.URL.Query().Get("offset"), "%d", &offset)
	fmt.Sscanf(r.URL.Query().Get("limit"), "%d", &limit)

	input := post.ListPostsInput{
		TeamID: teamID,
		UserID: userID,
		Offset: offset,
		Limit:  limit,
	}

	output, err := h.listPostsUC.Execute(r.Context(), input)
	if err != nil {
		respondError(w, http.StatusForbidden, err.Error())
		return
	}

	respondSuccess(w, output)
}

// ============================================================================
// Helper Functions (reuse from other handlers)
// ============================================================================

// func respondSuccess(w http.ResponseWriter, data interface{}) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteStatus(http.StatusOK)
// 	json.NewEncoder(w).Encode(data)
// }

// func respondCreated(w http.ResponseWriter, data interface{}) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(data)
// }

// func respondNoContent(w http.ResponseWriter) {
// 	w.WriteHeader(http.StatusNoContent)
// }

// func respondError(w http.ResponseWriter, code int, message string) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(code)
// 	json.NewEncoder(w).Encode(map[string]string{"error": message})
// }
