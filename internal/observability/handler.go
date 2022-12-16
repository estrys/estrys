package observability

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/gorilla/mux"
)

func HandleFunc(router *mux.Router, handler http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
			ctx = sentry.SetHubOnContext(ctx, hub)
		}
		options := []sentry.SpanOption{
			sentry.OpName("http.server"),
			sentry.ContinueFromRequest(request),
			sentry.TransctionSource(sentry.SourceRoute),
		}
		transactionName := fmt.Sprintf("%s %s", request.Method, request.URL.Path)
		var match mux.RouteMatch
		data := map[string]any{}
		router.Match(request, &match)
		if match.Route != nil {
			transactionName = match.Route.GetName()
			for k, v := range match.Vars {
				data[k] = v
			}
		}
		tx := StartTransaction(ctx, transactionName, options...)
		tx.Data = data

		defer FinishSpan(tx)
		request = request.WithContext(tx.Context()) //nolint:contextcheck
		hub.Scope().SetRequest(request)
		handler(writer, request)
	}
}
