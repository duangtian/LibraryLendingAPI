package middleware

import (
	"errors"
	"net/http"

	"github.com/example/librarylendingapi/internal/validation"
)

type appError struct {
	Status int
	Title string
	Err error
	Invalid []validation.InvalidParam
}

func (e *appError) Error() string { return e.Err.Error() }

func ProblemHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				validation.Write(w, http.StatusInternalServerError, "internal server error", "panic", nil)
			}
		}()
		rw := &respWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r)
		if rw.appErr != nil {
			e := rw.appErr
			validation.Write(w, e.Status, e.Title, e.Err.Error(), e.Invalid)
		}
	})
}

type respWriter struct {
	http.ResponseWriter
	appErr *appError
}

func (w *respWriter) WriteHeader(status int) { w.ResponseWriter.WriteHeader(status) }

func WriteError(w http.ResponseWriter, status int, title string, err error, invalid []validation.InvalidParam) {
	var ae *appError
	if errors.As(err, &ae) {
		w.WriteHeader(ae.Status)
		return
	}
	if rw, ok := w.(*respWriter); ok {
		rw.appErr = &appError{Status: status, Title: title, Err: err, Invalid: invalid}
	} else {
		validation.Write(w, status, title, err.Error(), invalid)
	}
}
