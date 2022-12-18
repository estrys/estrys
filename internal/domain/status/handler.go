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

	tweet, err := tweetRepository.GetTweet(request.Context(), tweetID)
	if err != nil {
		return internalerrors.Wrap(err, http.StatusNotFound).
			WithUserMessage("tweet not found")
	}

	if tweet.AuthorUsername != username {
		return internalerrors.Wrap(err, http.StatusBadRequest).
			WithUserMessage("tweet not found for this user")
	}

	user, err := userService.GetFullUser(request.Context(), tweet.AuthorUsername)
	if err != nil {
		return internalerrors.Wrap(err, http.StatusNotFound).
			WithUserMessage("user not found")
	}

	templateContent, _ := views.Views.ReadFile("status.html")
	statusTemplate := template.Must(template.New("status").Parse(string(templateContent)))
	responseWriter.Header().Add("content-type", "text/html")

	selfURL, err := urlGenerator.URL(
		routes.StatusRoute,
		[]string{"username", tweet.AuthorUsername, "id", tweetID},
		urlgenerator.OptionAbsoluteURL,
	)
	if err != nil {
		return internalerrors.Wrap(err, http.StatusInternalServerError).
			WithUserMessage("unable to generate status URL")
	}

	templateData := map[string]interface{}{
		"isRetweet": false,
		"url":       selfURL.String(),
		"tweet":     tweet,
		"user":      user,
	}
	// If we have a retweet, then fetch the retweet author and replace the current tweet by the retweet
	if retweet := tweet.Retweet(); retweet != nil {
		retweetUser, err := userService.GetFullUser(request.Context(), retweet.AuthorUsername)
		if err == nil {
			templateData["isRetweet"] = true
			templateData["originalUser"] = user
			templateData["tweet"] = retweet
			templateData["user"] = retweetUser
		}
	}
	err = statusTemplate.Execute(responseWriter, templateData)
	if err != nil {
		return internalerrors.Wrap(err, http.StatusInternalServerError)
	}
	return nil
}
