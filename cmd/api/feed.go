package main

import (
	"net/http"

	"github.com/truongndv/social/internal/store"
)

// getUserFeedHandler godoc
//
//	@Summary		Lấy feed của user
//	@Description	Danh sách bài viết kèm số comment, có phân trang, lọc và tìm kiếm (tạm thời gán cứng user ID = 1)
//	@Tags			feed
//	@Produce		json
//	@Param			limit	query		int		false	"Số bài tối đa (1-20)"	default(20)
//	@Param			offset	query		int		false	"Vị trí bắt đầu"		default(0)
//	@Param			sort	query		string	false	"Thứ tự theo thời gian"	Enums(asc, desc)	default(desc)
//	@Param			tags	query		string	false	"Lọc theo tag, cách nhau bằng dấu phẩy (tối đa 5)"
//	@Param			search	query		string	false	"Tìm theo title/content (tối đa 100 ký tự)"
//	@Param			since	query		string	false	"Từ thời điểm, định dạng 2006-01-02 15:04:05"
//	@Param			until	query		string	false	"Đến thời điểm, định dạng 2006-01-02 15:04:05"
//	@Success		200		{object}	DataResponse{data=[]store.PostWithMetadata}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		ApiKeyAuth
//	@Router			/users/feed [get]
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	userID := int64(1)

	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	feed, err := app.store.Posts.GetUserFeed(ctx, userID, fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, feed)
}
