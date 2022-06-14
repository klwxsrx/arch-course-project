package transport

import (
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"io/ioutil"
	"net/http"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	code int
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func NewLoggingMiddleware(l log.Logger, excludedURIs []string) func(next http.Handler) http.Handler {
	isExcluded := func(uri string) bool {
		for _, excluded := range excludedURIs {
			if uri == excluded {
				return true
			}
		}
		return false
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isExcluded(r.RequestURI) {
				next.ServeHTTP(w, r)
				return
			}

			lrw := newLoggingResponseWriter(w)
			next.ServeHTTP(lrw, r)

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				body = nil
			}
			loggerWithFields := l.With(log.Fields{
				"method":       r.Method,
				"url":          r.RequestURI,
				"body":         string(body),
				"responseCode": lrw.code,
			})
			if lrw.code == http.StatusInternalServerError {
				loggerWithFields.Error("internal server error")
			} else {
				loggerWithFields.Info("request handled")
			}
		})
	}
}
