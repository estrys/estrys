package router

import (
	"github.com/gorilla/mux"
	"net/http"

	"github.com/estrys/estrys/internal/activitypub/routes"
	"github.com/estrys/estrys/internal/domain/status"
)

func GetRouter() *mux.Router {
	r := mux.NewRouter()
	routes.Router(r)
	r.Path("/").HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Header().Add("Location", "https://github.com/estrys/estrys")
		responseWriter.WriteHeader(http.StatusFound)
	})
	status.StatusRouter(r)
	return r
}
