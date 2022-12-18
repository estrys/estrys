package logger

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/estrys/estrys/internal/observability"
)

type HTTPLoggerRoundTripper struct {
	RoundTripper http.RoundTripper
	Log          Logger
}

func (h *HTTPLoggerRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if request != nil {
		span := observability.StartSpan(
			request.Context(),
			"http.client",
			map[string]any{"http.url": request.URL.String()},
		)
		resp, err := h.RoundTripper.RoundTrip(request)
		observability.FinishSpan(span)
		log := h.Log.WithFields(logrus.Fields{
			"host":   request.Host,
			"method": request.Method,
			"url":    request.URL.Path,
			"query":  request.URL.Query(),
		})
		if resp != nil {
			log = log.WithField("status", resp.StatusCode)
		}
		if err != nil {
			log = log.WithError(err)
		}
		log.Trace("http call")
		return resp, err
	}
	return h.RoundTripper.RoundTrip(request)
}
