package urlgenerator

import (
	"net/url"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/config"
)

type GenerateRouteOptions int

const (
	OptionAbsoluteURL GenerateRouteOptions = iota
)

type RouteParams []string

type URLGenerator interface {
	URL(string, RouteParams, ...GenerateRouteOptions) (*url.URL, error)
}

type muxURLGenerator struct {
	config config.Config
	router *mux.Router
}

func NewURLGenerator(config config.Config, router *mux.Router) *muxURLGenerator {
	return &muxURLGenerator{
		config: config,
		router: router,
	}
}

func (r *muxURLGenerator) URL(routeName string, params RouteParams, options ...GenerateRouteOptions) (*url.URL, error) {
	routeURL, err := r.router.Get(routeName).URL(params...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate URL for route")
	}
	for _, option := range options {
		if option == OptionAbsoluteURL {
			routeURL.Host = r.config.Domain.Host
			routeURL.Scheme = r.config.Domain.Scheme
		}
	}
	return routeURL, nil
}
