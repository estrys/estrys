package logger

import (
	"net/http"

	"github.com/motemen/go-loghttp"
	"github.com/sirupsen/logrus"
)

func GetResponseLogger(log Logger) *loghttp.Transport {
	return &loghttp.Transport{
		LogRequest: func(req *http.Request) {},
		LogResponse: func(resp *http.Response) {
			log.WithFields(logrus.Fields{
				"host":   resp.Request.Host,
				"method": resp.Request.Method,
				"status": resp.StatusCode,
				"url":    resp.Request.URL.Path,
				"query":  resp.Request.URL.Query(),
			}).Trace("http call")
		},
	}
}
