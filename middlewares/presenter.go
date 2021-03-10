package middlewares

import (
	"github.com/DmitriBeattie/custom-framework/interfaces/api"
	"github.com/gorilla/context"
	"net/http"
)

type httpPresenter struct {
	api.ErrorPresenter
	api.ResponsePresenter
	api.RequestDataParser
}


func PresenterMiddleware(errPr api.ErrorPresenter, respPr api.ResponsePresenter, reqPr api.RequestDataParser) api.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			presenter := httpPresenter{
				ErrorPresenter:    errPr,
				ResponsePresenter: respPr,
				RequestDataParser: reqPr,
			}

			context.Set(r, "presenter", presenter)

			next(w, r)
		}
	}
}

