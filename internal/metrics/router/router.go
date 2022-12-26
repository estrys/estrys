package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/metrics"
	"github.com/estrys/estrys/internal/router/routes"
)

func Router(rootRouter *mux.Router) {
	meter := dic.GetService[metrics.Meter]()
	reg := meter.GetRegistry()
	userRouter := rootRouter.PathPrefix("/metrics").Subrouter()
	userRouter.NewRoute().Name(routes.MetricsRoute).
		Path("").
		Methods(http.MethodGet).
		Handler(promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
}
