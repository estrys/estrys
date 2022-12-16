package status

import (
	"net/http"
	"text/template"

	"github.com/gorilla/mux"

	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/domain"
	"github.com/estrys/estrys/internal/domain/status/views"
	internalerrors "github.com/estrys/estrys/internal/errors"
	"github.com/estrys/estrys/internal/router/routes"
	"github.com/estrys/estrys/internal/router/urlgenerator"
	"github.com/estrys/estrys/internal/twitter/repository"
)

func HandleStatus(responseWriter http.ResponseWriter, request *http.Request) error {
	vars := mux.Vars(request)
	tweetRepository := dic.GetService[repository.TweetRepository]()
	userService := dic.GetService[domain.UserService]()
	urlGenerator := dic.GetService[urlgenerator.URLGenerator]()

	username := vars["username"]
	tweetID := vars["id"]
	if username == "" || tweetID == "" {
		return internalerrors.New("either username or id is not set", http.StatusBadRequest)
	}

	user, err := userService.GetFullUser(request.Context(), username)
	if err != nil {
		return internalerrors.Wrap(err, http.StatusNotFound).
			WithUserMessage("user not found")
	}

	tweet, err := tweetRepository.GetTweet(request.Context(), tweetID)
	if err != nil {
		return internalerrors.Wrap(err, http.StatusNotFound).
			WithUserMessage("tweet not found")
	}

	templateContent, _ := views.Views.ReadFile("status.html")
	statusTemplate := template.Must(template.New("status").Parse(string(templateContent)))
	responseWriter.Header().Add("content-type", "text/html")

	selfURL, err := urlGenerator.URL(
		routes.StatusRoute,
		[]string{"username", username, "id", tweetID},
		urlgenerator.OptionAbsoluteURL,
	)
	if err != nil {
		return internalerrors.Wrap(err, http.StatusInternalServerError).
			WithUserMessage("unable to generate status URL")
	}

	_ = statusTemplate.Execute(responseWriter, map[string]interface{}{
		"url":   selfURL.String(),
		"tweet": tweet,
		"user":  user,
	})
	return nil
}
