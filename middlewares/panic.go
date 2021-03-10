package middlewares

import (
	"github.com/DmitriBeattie/custom-framework/abstract/apierror"
	"github.com/DmitriBeattie/custom-framework/interfaces/api"
	"github.com/DmitriBeattie/custom-framework/interfaces/app"
	"fmt"
	"net/http"
	"runtime/debug"
)

func InternalError() apierror.APIError {
	return apierror.New().
		FromMsg("Internal Server Error").
		Component("middlewares").
		Code("PanicRecoveryMiddleware")
}

func PanicRecoveryMiddleware(logger app.Logger, pr api.ErrorPresenter) api.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {

					err := fmt.Errorf("Ошибка %s. %s", fmt.Sprint(rec), string(debug.Stack()))

					logger.Error(err)

					pr.Error(
						w,
						r,
						InternalError(),
						http.StatusInternalServerError,
						nil,
					)
				}
			}()

			next(w, r)
		}
	}
}
