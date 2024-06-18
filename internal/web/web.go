package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var cookieExp = 24 * time.Hour

func SendError(w http.ResponseWriter, msg string, code int) {
	fmt.Printf("%d error: %s\n", code, msg)
	if code >= 500 {
		msg = "something went wrong"
	}
	http.Error(w, msg, code)
}

func InternalError(w http.ResponseWriter, err error) {
	SendError(w, err.Error(), http.StatusInternalServerError)
}

func NotFound(w http.ResponseWriter, entity string, id string) {
	SendError(w, fmt.Sprintf("%s %s not found", entity, id), http.StatusNotFound)
}

func TooManyRequests(w http.ResponseWriter, msg string) {
	SendError(w, msg, http.StatusTooManyRequests)
}

func BadRequest(w http.ResponseWriter, err error) {
	SendError(w, err.Error(), http.StatusBadRequest)
}

func SetCookie(w http.ResponseWriter, name string, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		Expires:  time.Now().Add(cookieExp),
	})
}

func DecodeBody[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	b := new(T)
	err := json.NewDecoder(r.Body).Decode(b)
	if err != nil {
		BadRequest(w, err)
		return *b, false
	}
	return *b, true
}

func JSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
