package middlewares

import (
	"delivery-report/framework/abstract/apierror"
	"delivery-report/framework/interfaces/api"
	"delivery-report/framework/interfaces/app"
	"net/http"

	"github.com/gorilla/context"
	"github.com/sarulabs/di"
)

func DependencyInjectionError() apierror.APIError {
	return apierror.New().
		Component("middlewares").
		FromMsg("Не удалось удалить контейнер из памяти").
		Code("DependencyInjectionMiddleware")
}

func DependencyInjectionMiddleware(appCtn di.Container, l app.Logger) api.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctn, err := appCtn.SubContainer()
			if err != nil {
				panic(err)
			}
			defer func() {
				if err := ctn.Delete(); err != nil {
					l.Error(DependencyInjectionError)
				}
			}()

			context.Set(r, di.ContainerKey("di"), ctn)

			next(w, r)
		}
	}
}
