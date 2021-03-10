package middlewares

import (
	"github.com/DmitriBeattie/custom-framework/interfaces/api"
	"github.com/DmitriBeattie/custom-framework/interfaces/translator"
	"net/http"

	"github.com/gorilla/context"
)

const DEFAULTLANGUAGE translator.Language = "ru"

func LocaleMiddleware() api.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var language translator.Language

			param := r.Header.Get("Accept-Language")
			if param == "" {
				language = DEFAULTLANGUAGE
			} else {
				language = translator.Language(param)
			}

			context.Set(r, "lang", language)

			next(w, r)
		}
	}
}
