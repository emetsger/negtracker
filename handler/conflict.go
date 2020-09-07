package handler

import (
	"net/http"
	"strconv"
)

func Conflict(w http.ResponseWriter, r *http.Request, reason string) {
	bytes := []byte(reason)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
	w.WriteHeader(409)
	_, _ = w.Write(bytes)
}
