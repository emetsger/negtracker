package neg

import (
	"bytes"
	"encoding/json"
	"github.com/emetsger/negtracker/handler"
	"github.com/emetsger/negtracker/model"
	"github.com/emetsger/negtracker/store"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var negHandler = func(w http.ResponseWriter, r *http.Request) {
	var h http.HandlerFunc
	switch r.Method {
	case http.MethodGet:
		h = wrap([]byte("Placeholder for returned Neg by ID"), 200, "text/plain", r, w)
	case http.MethodPost:
		h = wrap([]byte("Placeholder for creating a neg record"), 201, "text/plain", r, w)
	default:
		h = func(w http.ResponseWriter, r *http.Request) {
			handler.NotImplemented(w, r)
		}
	}

	h.ServeHTTP(w, r)
}

func NewHandler(store store.Api) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var h http.HandlerFunc
		switch r.Method {
		case http.MethodGet:
			if id := parseIdFromUri(r.URL.String()); id == "" {
				// id could not be parsed from the URI
				h = func(w http.ResponseWriter, r *http.Request) {
					handler.MalformedRequest(w, r)
				}
			} else {
				neg := model.Neg{}
				if err := store.Retrieve(id, &neg); err != nil {
					h = func(w http.ResponseWriter, r *http.Request) {
						handler.ServerError(w, r)
					}
				} else {
					if body, err := json.Marshal(neg); err != nil {
						h = func(w http.ResponseWriter, r *http.Request) {
							handler.ServerError(w, r)
						}
					} else {
						h = wrap(body, 200, "application/json", r, w)
					}
				}
			}
		case http.MethodPost:
			buf := &bytes.Buffer{}
			_, _ = io.Copy(buf, r.Body)
			if buf.Len() < 1 {
				// no request body, nothing to create
				h = func(w http.ResponseWriter, r *http.Request) {
					handler.MalformedRequest(w, r)
				}
			} else {
				n := &model.Neg{}
				if err := json.Unmarshal(buf.Bytes(), n); err != nil {
					// malformed body
					h = func(w http.ResponseWriter, r *http.Request) {
						handler.MalformedRequest(w, r)
					}
				} else {
					if id, err := store.Store(*n); err != nil {
						// error storing the neg
						h = func(w http.ResponseWriter, r *http.Request) {
							handler.ServerError(w, r)
						}
					} else {
						// return a 201 TODO decide on id approach
						h = wrap([]byte("TODO URL "+id), 201, "text/plain", r, w)
					}
				}
			}
		default:
			h = func(w http.ResponseWriter, r *http.Request) {
				handler.NotImplemented(w, r)
			}
		}

		h.ServeHTTP(w, r)
	}
}

// TODO: test, e.g., when the parsed id is not valid, things panic in the store layer, and empty response is returned
func parseIdFromUri(uri string) string {
	index := strings.LastIndex(uri, "/")
	if index > -1 && index < len(uri) {
		return uri[index+1:]
	}

	return ""
}

func wrap(body []byte, status int, mediaType string, r *http.Request, w http.ResponseWriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.Header().Set("Content-Type", mediaType)

		if status > 199 && status < 600 {
			w.WriteHeader(status)
		}

		if len(body) > 0 {
			_, _ = w.Write(body)
		}
	}
}
