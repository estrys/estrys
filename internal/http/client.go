package http

import "net/http"

//go:generate mockery --with-expecter --name=Client
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}
