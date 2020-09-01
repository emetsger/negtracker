package neg

import (
	"github.com/emetsger/negtracker/handler"
	"net/http"
	"strconv"
)

var NegHandler = func(w http.ResponseWriter, r *http.Request) {
	var h http.HandlerFunc
	switch r.Method {
	case http.MethodGet:
		h = wrap([]byte("Placeholder for returned Neg by ID"), 200, w, r)
	case http.MethodPost:
		h = wrap([]byte("Placeholder for creating a neg record"), 201, w, r)
	default:
		h = func(w http.ResponseWriter, r *http.Request) {
			handler.NotImplemented(w, r)
		}
	}

	h.ServeHTTP(w, r)
}

func wrap(body []byte, status int, w http.ResponseWriter, r *http.Request) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.Header().Set("Content-Type", "text/plain")

		if status > 199 && status < 600 {
			w.WriteHeader(status)
		}

		if len(body) > 0 {
			_, _ = w.Write(body)
		}
	}
}

