package router

import (
	"github.com/gorilla/mux"

	"github.com/estrys/estrys/internal/activitypub/routes"
	"github.com/estrys/estrys/internal/domain/status"
)

func GetRouter() *mux.Router {
	r := mux.NewRouter()
	routes.Router(r)
	status.StatusRouter(r)
	return r
}
