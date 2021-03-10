package middlewares

import (
	"fmt"
	"delivery-report/framework/interfaces/api"
	"delivery-report/framework/interfaces/app"
	"net/http"
	"runtime/debug"
)

func PanicRecoveryMiddlewar_v2(logger app.Logger, pr api.ErrorPresenter) api.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {

					err := fmt.Errorf("Ошибка %s. %s", fmt.Sprint(rec), string(debug.Stack()))

					logger.Error(err)

					pr.Error(
						w,
						r,
						err,
						http.StatusInternalServerError,
						nil,
					)
				}
			}()

			next(w, r)
		}
	}
}

