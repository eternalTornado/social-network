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

// createPostHandler godoc
//
//	@Summary		Tạo bài viết
//	@Description	Tạo bài viết mới (tạm thời gán cứng user ID = 1, chưa có đăng nhập)
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreatePostPayload	true	"Nội dung bài viết"
//	@Success		201		{object}	DataResponse{data=store.Post}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
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

// getPostHandler godoc
//
//	@Summary		Lấy chi tiết bài viết
//	@Description	Lấy bài viết theo ID, kèm danh sách comment
//	@Tags			posts
//	@Produce		json
//	@Param			postID	path		int	true	"Post ID"
//	@Success		200		{object}	DataResponse{data=store.Post}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID} [get]
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

// updatePostHandler godoc
//
//	@Summary		Cập nhật bài viết
//	@Description	Cập nhật một phần (PATCH) title và/hoặc content; field bỏ trống thì giữ nguyên
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int					true	"Post ID"
//	@Param			payload	body		UpdatePostPayload	true	"Các field cần cập nhật"
//	@Success		200		{object}	DataResponse{data=store.Post}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID} [patch]
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

// deletePostHandler godoc
//
//	@Summary	Xoá bài viết
//	@Tags		posts
//	@Param		postID	path	int	true	"Post ID"
//	@Success	204		"No Content"
//	@Failure	400		{object}	ErrorResponse
//	@Failure	404		{object}	ErrorResponse
//	@Failure	500		{object}	ErrorResponse
//	@Security	ApiKeyAuth
//	@Router		/posts/{postID} [delete]
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

// createCommentHandler godoc
//
//	@Summary		Thêm comment vào bài viết
//	@Description	Tạo comment cho bài viết (tạm thời gán cứng user ID = 1, chưa có đăng nhập)
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int					true	"Post ID"
//	@Param			payload	body		PostCommentsPayload	true	"Nội dung comment"
//	@Success		201		{object}	DataResponse{data=store.Comment}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID}/comments [post]
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
