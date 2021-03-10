package http

import (
	"encoding/json"
	"net/http"
)

var SimplePresenterInstance = PresenterInstance{}

type PresenterInstance struct{}

func (rInstance PresenterInstance) Response(w http.ResponseWriter, r *http.Request, data interface{}) {
	ResponseModel(w, data)
}

func (rInstance PresenterInstance) Error(w http.ResponseWriter, r *http.Request, err error, code int) {
	ResponseWithError(w, err, code)
}

func setHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
	setHeader(w)
}

func ResponseModel(w http.ResponseWriter, respModel interface{}) {
	setHeader(w)
	json.NewEncoder(w).Encode(respModel)
}

func ResponseWithError(w http.ResponseWriter, err error, errorCode int) {
	errData := struct {
		ErrorString string `json:"error"`
		ErrorCode   int    `json:"code"`
	}{
		ErrorString: err.Error(),
		ErrorCode:   errorCode,
	}

	setHeader(w)
	w.WriteHeader(errorCode)

	json.NewEncoder(w).Encode(errData)
}