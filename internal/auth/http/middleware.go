package http

import (
	"julo/internal/auth"
	httphelper "julo/internal/http"
	"log"
	"net/http"
	"strings"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		_, token, found := strings.Cut(header, "Token ")
		if !found || token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		session, err := auth.GetSession(r.Context(), token)
		if err != nil {
			switch err {
			case auth.ErrSessionNotFound:
				httphelper.WriteErrorJSON(w, http.StatusUnauthorized, err)
				return
			default:
				log.Println(err)
				httphelper.WriteErrorJSON(w, http.StatusInternalServerError, err)
				return
			}
		}

		if session == nil {
			httphelper.WriteErrorJSON(w, http.StatusUnauthorized, auth.ErrSessionNotFound)
			return
		}

		c := auth.SessionIntoContext(r.Context(), session)
		next.ServeHTTP(w, r.WithContext(c))
	})
}
