package router

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/estrys/estrys/internal/activitypub/routes"
	"github.com/estrys/estrys/internal/domain/status"
	metrics "github.com/estrys/estrys/internal/metrics/router"
)

func GetRouter() *mux.Router {
	newRouter := mux.NewRouter()
	routes.Router(newRouter)
	metrics.Router(newRouter)
	status.StatusRouter(newRouter)
	newRouter.Path("/").
		HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.Header().Add("Location", "https://github.com/estrys/estrys")
			responseWriter.WriteHeader(http.StatusFound)
		})
	return newRouter
}
