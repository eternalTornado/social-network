package main

import (
	"log"
	"net/http"
)

func internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal error: method %s path %s error %s", r.Method, r.URL.Path, err.Error())

	writeJSONError(w, http.StatusInternalServerError, "the server encoutered a problem and cannot process the request")
}

func badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request: method %s path %s error %s", r.Method, r.URL.Path, err.Error())

	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found: method %s path %s error %s", r.Method, r.URL.Path, err.Error())
	writeJSONError(w, http.StatusNotFound, "not found")
}
