package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_576 // không cho phép body quá 1MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

// DataResponse là envelope cho mọi response thành công: {"data": ...}
type DataResponse struct {
	Data any `json:"data"`
}

// ErrorResponse là envelope cho mọi response lỗi: {"Error": "..."}
type ErrorResponse struct {
	Error string `json:"Error"`
}

func writeJSONError(w http.ResponseWriter, status int, message string) error {
	return writeJSON(w, status, &ErrorResponse{Error: message})
}

func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	return writeJSON(w, status, &DataResponse{Data: data})
}
