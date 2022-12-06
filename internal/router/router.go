package router

import (
	"github.com/gorilla/mux"

	"github.com/estrys/estrys/internal/activitypub/routes"
)

func GetRouter() *mux.Router {
	r := mux.NewRouter()
	routes.Router(r)
	return r
}
