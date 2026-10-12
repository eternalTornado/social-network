package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/truongndv/social/internal/store"
)

type postKey string

const postCtx postKey = "post"

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

	//writeJSON(w, http.StatusCreated, post)
	app.jsonResponse(w, http.StatusCreated, post)
}

// GET /v1/posts/{postID}
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	ctx := r.Context()

	comments, err := app.store.Comments.GetByPostID(ctx, post.ID)
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	post.Comments = comments

	//writeJSON(w, http.StatusOK, post)
	app.jsonResponse(w, http.StatusOK, post)
}

// Note: lí do dùng *string là để phân biệt giữa nil và ""
// nil -> bad request
// "" -> update value thành ""
type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=100"`
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	var payload UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		badRequestError(w, r, err)
		return
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}

	ctx := r.Context()

	if err := app.store.Posts.Update(ctx, post); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			notFoundError(w, r, err)
		default:
			internalServerError(w, r, err)
		}

		return
	}

	//writeJSON(w, http.StatusOK, post)
	app.jsonResponse(w, http.StatusOK, post)
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	ctx := r.Context()

	if err := app.store.Posts.Delete(ctx, post.ID); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			notFoundError(w, r, err)
		default:
			internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "postID")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			badRequestError(w, r, err)
			return
		}

		ctx := r.Context()
		post, err := app.store.Posts.GetByID(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				notFoundError(w, r, err)
			default:
				internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {

	value, ok := r.Context().Value(postCtx).(*store.Post)
	if !ok {
		return nil
	}
	return value
}

type PostCommentsPayload struct {
	Content string `json:"content" validate:"required,max=1000"`
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {

	var payload PostCommentsPayload
	err := readJSON(w, r, &payload)
	if err != nil {
		badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		badRequestError(w, r, err)
		return
	}

	post := getPostFromCtx(r)

	newComment := &store.Comment{
		PostID:  post.ID,
		UserID:  1,
		Content: payload.Content,
	}
	ctx := r.Context()

	if err := app.store.Comments.Create(ctx, newComment); err != nil {
		internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusCreated, newComment)
}
