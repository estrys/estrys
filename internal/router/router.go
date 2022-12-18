package router

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/estrys/estrys/internal/activitypub/routes"
	"github.com/estrys/estrys/internal/domain/status"
)

func GetRouter() *mux.Router {
	newRouter := mux.NewRouter()
	routes.Router(newRouter)
	newRouter.Path("/").HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Header().Add("Location", "https://github.com/estrys/estrys")
		responseWriter.WriteHeader(http.StatusFound)
	})
	status.StatusRouter(newRouter)
	return newRouter
}
