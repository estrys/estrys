package routes

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/estrys/estrys/internal/activitypub/auth"
	"github.com/estrys/estrys/internal/activitypub/handlers"
	"github.com/estrys/estrys/internal/router/routes"
)

func Router(rootRouter *mux.Router) {
	wellKnownRouter := rootRouter.PathPrefix("/.well-known").Subrouter()
	wellKnownRouter.NewRoute().Name(routes.HostMetaRoute).
		Path("/host-meta").
		Methods(http.MethodGet).
		HandlerFunc(handlers.HandleHostMeta)
	wellKnownRouter.NewRoute().Name(routes.WebfingerRoute).
		Path("/webfinger").
		Methods(http.MethodGet).
		HandlerFunc(handlers.HandleWebFinger)

	// TODO Add regex on Accept header
	userRouter := rootRouter.PathPrefix("/users").Subrouter()
	userRouter.NewRoute().Name(routes.UserRoute).
		Path("/{username}").
		Methods(http.MethodGet).
		HandlerFunc(handlers.HandleUser)
	userRouter.NewRoute().Name(routes.UserFollowingRoute).
		Path("/{username}/following").
		Methods(http.MethodGet).
		HandlerFunc(handlers.HandleFollowing)
	userRouter.NewRoute().Name(routes.UserFollowersRoute).
		Path("/{username}/followers").
		Methods(http.MethodGet).
		HandlerFunc(handlers.HandleFollowers)
	userRouter.NewRoute().Name(routes.UserOutbox).
		Path("/{username}/outbox").
		Methods(http.MethodGet).
		HandlerFunc(handlers.HandleOutbox)

	inboxRouter := userRouter.PathPrefix("/{username}/inbox").Subrouter()
	inboxRouter.Use(auth.HTTPSigMiddleware)
	inboxRouter.NewRoute().Name(routes.UserInbox).
		Path("").
		Methods(http.MethodPost).
		HandlerFunc(handlers.HandleInbox)
}
