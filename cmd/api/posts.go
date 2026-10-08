package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/truongndv/social/internal/store"
)

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=100"`
	Tags    []string `json:"tags"`
}

// POST /v1/posts
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		badRequestError(w, r, err)
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		ID:      1, // tạm gán, chưa đăng nhập
	}

	ctx := r.Context()

	if err := app.store.Posts.Create(ctx, post); err != nil {
		internalServerError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, post)
}

// GET /v1/posts/{postID}
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	paramStr := chi.URLParam(r, "postID")
	postID, err := strconv.Atoi(paramStr)
	if err != nil {
		badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	post, err := app.store.Posts.GetByID(ctx, int64(postID))
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			notFoundError(w, r, err)
		default:
			internalServerError(w, r, err)
		}
		return
	}

	comments, err := app.store.Comments.GetByPostID(ctx, post.ID)
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	post.Comments = comments

	writeJSON(w, http.StatusOK, post)
}
