package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/truongndv/social/internal/store"
)

type userKey string

const userCtx userKey = "user"

type FollowUser struct {
	UserID int64 `json:"user_id"`
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	app.jsonResponse(w, http.StatusOK, user)
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followedUser := getUserFromContext(r)

	var payload FollowUser
	if err := readJSON(w, r, &payload); err != nil {
		badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	if err := app.store.Followers.Follow(ctx, payload.UserID, followedUser.ID); err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			conflictError(w, r, err)
		default:
			internalServerError(w, r, err)
		}
		return
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followedUser := getUserFromContext(r)

	var payload FollowUser
	if err := readJSON(w, r, &payload); err != nil {
		badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	if err := app.store.Followers.Unfollow(ctx, payload.UserID, followedUser.ID); err != nil {
		internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "userID")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			badRequestError(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := app.store.Users.GetByID(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				notFoundError(w, r, err)
			default:
				internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, userCtx, *user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromContext(r *http.Request) *store.User {
	user, ok := r.Context().Value(userCtx).(*store.User)
	if !ok {
		return nil
	}

	return user
}
