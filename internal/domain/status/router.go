package status

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/estrys/estrys/internal/errors"
	"github.com/estrys/estrys/internal/router/routes"
)

func StatusRouter(rootRouter *mux.Router) {
	userRouter := rootRouter.PathPrefix("/status").Subrouter()
	userRouter.NewRoute().Name(routes.StatusRoute).
		Path("/{username}/{id}").
		Methods(http.MethodGet).
		HandlerFunc(errors.HTTPErrorHandler(HandleStatus))
}
