package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	//app.logger.Errorf("internal error: method %s path %s error %s", r.Method, r.URL.Path, err.Error())
	app.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusInternalServerError, "the server encoutered a problem and cannot process the request")
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	//log.Printf("bad request: method %s path %s error %s", r.Method, r.URL.Path, err.Error())
	app.logger.Warnw("bad request", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	//log.Printf("not found: method %s path %s error %s", r.Method, r.URL.Path, err.Error())
	app.logger.Warnw("not found", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusNotFound, "not found")
}

func (app *application) conflictError(w http.ResponseWriter, r *http.Request, err error) {
	//log.Printf("conflict: method %s path %s error %s", r.Method, r.URL.Path, err.Error())
	app.logger.Warnw("conflict", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusConflict, err.Error())
}
