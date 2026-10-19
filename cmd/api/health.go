package main

import "net/http"

// healthCheckHandler godoc
//
//	@Summary		Kiểm tra tình trạng server
//	@Description	Trả về trạng thái, môi trường và version của API
//	@Tags			ops
//	@Produce		json
//	@Success		200	{object}	DataResponse{data=map[string]string}
//	@Failure		500	{object}	ErrorResponse
//	@Security		ApiKeyAuth
//	@Router			/health [get]
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": version,
	}

	if err := app.jsonResponse(w, http.StatusOK, data); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
}
