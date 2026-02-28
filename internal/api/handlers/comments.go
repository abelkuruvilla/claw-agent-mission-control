package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/google/uuid"

	"github.com/abelkuruvilla/claw-agent-mission-control/internal/db"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/store"
)

type CommentHandler struct {
	store *store.Store
}

func NewCommentHandler(s *store.Store) *CommentHandler {
	return &CommentHandler{
		store: s,
	}
}

// Request types
type CreateCommentRequest struct {
	Author  string `json:"author" validate:"required"`
	Content string `json:"content" validate:"required"`
}

// Response types
type CommentResponse struct {
	ID        string `json:"id"`
	TaskID    string `json:"task_id"`
	Author    string `json:"author"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// List all comments for a task
func (h *CommentHandler) ListByTask(c echo.Context) error {
	taskID := c.Param("id")

	// Verify task exists
	_, err := h.store.GetTask(c.Request().Context(), taskID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Task not found")
	}

	comments, err := h.store.ListCommentsByTask(c.Request().Context(), taskID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	responses := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = toCommentResponse(comment)
	}

	return c.JSON(http.StatusOK, responses)
}

// Create a new comment for a task
func (h *CommentHandler) Create(c echo.Context) error {
	taskID := c.Param("id")
	
	// Verify task exists
	_, err := h.store.GetTask(c.Request().Context(), taskID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Task not found")
	}

	var req CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Generate UUID for new comment
	id := uuid.New().String()

	comment, err := h.store.CreateComment(c.Request().Context(), db.CreateCommentParams{
		ID:      id,
		TaskID:  taskID,
		Author:  req.Author,
		Content: req.Content,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, toCommentResponse(comment))
}

// Delete a comment
func (h *CommentHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	
	// Verify comment exists
	_, err := h.store.GetComment(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Comment not found")
	}

	if err := h.store.DeleteComment(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	
	return c.NoContent(http.StatusNoContent)
}

// Helper functions
func toCommentResponse(comment db.Comment) CommentResponse {
	return CommentResponse{
		ID:        comment.ID,
		TaskID:    comment.TaskID,
		Author:    comment.Author,
		Content:   comment.Content,
		CreatedAt: nullTimeToString(comment.CreatedAt),
	}
}
