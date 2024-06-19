package webutils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Web struct {
	cookieExp time.Duration
	logger    *slog.Logger
}

func New(cookieExp time.Duration, logger *slog.Logger) Web {
	return Web{
		cookieExp: cookieExp,
		logger:    logger,
	}
}

func (web Web) SendError(w http.ResponseWriter, msg string, code int) {
	if code >= 500 {
		web.logger.Error("sending HTTP error", "code", code, "err", msg)
		http.Error(w, "something went wrong", code)
	} else {
		web.logger.Info("sending HTTP error", "code", code, "err", msg)
		http.Error(w, msg, code)
	}
}

func (web Web) InternalError(w http.ResponseWriter, err error) {
	web.SendError(w, err.Error(), http.StatusInternalServerError)
}

func (web Web) NotFound(w http.ResponseWriter, entity string, id string) {
	web.SendError(w, fmt.Sprintf("%s %s not found", entity, id), http.StatusNotFound)
}

func (web Web) TooManyRequests(w http.ResponseWriter, msg string) {
	web.SendError(w, msg, http.StatusTooManyRequests)
}

func (web Web) BadRequest(w http.ResponseWriter, err error) {
	web.SendError(w, err.Error(), http.StatusBadRequest)
}

func (web Web) SetCookie(w http.ResponseWriter, name string, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		Expires:  time.Now().Add(web.cookieExp),
	})
}

func (web Web) DecodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		web.BadRequest(w, err)
		return false
	}
	return true
}

func (web Web) JSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		web.InternalError(w, err)
	}
}
