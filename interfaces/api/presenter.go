package api

import (
	"delivery-report/framework/abstract/apierror"
	"encoding/json"
	"net/http"
)

type HTTPPresenter interface {
	ErrorPresenter
	RequestDataParser
	ResponsePresenter
}

type ErrorPresenter interface {
	Error(w http.ResponseWriter, r *http.Request, err error, code int, extInfo interface{})
}

type RequestDataParser interface {
	ParseRequest(r *http.Request) (interface{}, error)
}

type ResponsePresenter interface {
	Response(w http.ResponseWriter, r *http.Request, data interface{})
}

type defaultPresenter struct{}

var DefaultPresenter defaultPresenter

type presentError struct {
	ErrMsg         string          `json:"err_msg"`
	Code           int             `json:"code"`
	ErrCode        string          `json:"internal_code,omitempty"`
	AdditionalInfo json.RawMessage `json:"ext_info,omitempty"`
}

func setHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}


func (d defaultPresenter) Error(w http.ResponseWriter, r *http.Request, err error, code int, extInfo interface{}) {
	if code == 0 {
		code = http.StatusInternalServerError
	}

	var pE presentError
	pE.Code = code
	pE.ErrMsg = err.Error()

	if e, ok := err.(apierror.APIError); ok {
		pE.ErrCode = e.ID()
	}

	if extInfo != nil {
		pE.AdditionalInfo, _ = json.Marshal(extInfo)
	}

	setHeader(w)
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(pE)
}


func (d defaultPresenter) Response(w http.ResponseWriter, r *http.Request, data interface{})  {
	setHeader(w)
	json.NewEncoder(w).Encode(data)
}