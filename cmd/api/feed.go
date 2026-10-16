package main

import (
	"net/http"

	"github.com/truongndv/social/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	userID := int64(1)

	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	feed, err := app.store.Posts.GetUserFeed(ctx, userID, fq)
	if err != nil {
		internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, feed)
}
