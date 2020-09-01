package handler

import (
	"net/http"
	"strconv"
)

func NotImplemented(w http.ResponseWriter, r *http.Request) {
	bytes := []byte("Method not implemented")
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
	w.WriteHeader(500)
	_, _ = w.Write(bytes)
}