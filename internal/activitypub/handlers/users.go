package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-fed/activity/streams"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/activitypub"
	"github.com/estrys/estrys/internal/activitypub/auth"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/domain"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/twitter"
)

func HandleUser(responseWriter http.ResponseWriter, request *http.Request) {
	log := dic.GetService[logger.Logger]()

	vars := mux.Vars(request)
	// TODO Validate username input

	userService := dic.GetService[domain.UserService]()
	user, err := userService.GetFullUser(request.Context(), vars["username"])

	if err != nil {
		var twitterUserNotFound twitter.UsernameNotFoundError
		if errors.As(err, &twitterUserNotFound) || errors.Is(err, domain.ErrUserDoesNotExist) {
			responseWriter.WriteHeader(http.StatusNotFound)
		} else {
			responseWriter.WriteHeader(http.StatusInternalServerError)
		}
		log.WithError(err).Error("unable to retrieve user")
		return
	}

	vocabService := dic.GetService[activitypub.VocabService]()
	actor, err := vocabService.GetActor(user)
	if err != nil {
		log.Error(err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Add("content-type", "application/activity+json")
	err = json.NewEncoder(responseWriter).Encode(actor)
	if err != nil {
		log.Error(err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func HandleFollowing(responseWriter http.ResponseWriter, request *http.Request) {
	log := dic.GetService[logger.Logger]()

	vars := mux.Vars(request)
	userService := dic.GetService[domain.UserService]()
	user, err := userService.GetFullUser(request.Context(), vars["username"])

	if err != nil {
		var twitterUserNotFound twitter.UsernameNotFoundError
		if errors.As(err, &twitterUserNotFound) || errors.Is(err, domain.ErrUserDoesNotExist) {
			responseWriter.WriteHeader(http.StatusNotFound)
		} else {
			responseWriter.WriteHeader(http.StatusInternalServerError)
		}
		log.WithError(err).Error("unable to retrieve user")
		return
	}

	vocabService := dic.GetService[activitypub.VocabService]()
	following, err := vocabService.GetFollowing(user)
	if err != nil {
		log.Error(err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Add("content-type", "application/activity+json")
	err = json.NewEncoder(responseWriter).Encode(following)
	if err != nil {
		log.Error(err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func HandleFollowers(responseWriter http.ResponseWriter, request *http.Request) {
	log := dic.GetService[logger.Logger]()

	vars := mux.Vars(request)
	userService := dic.GetService[domain.UserService]()
	user, err := userService.GetFullUser(request.Context(), vars["username"])

	if err != nil {
		var twitterUserNotFound twitter.UsernameNotFoundError
		if errors.As(err, &twitterUserNotFound) || errors.Is(err, domain.ErrUserDoesNotExist) {
			responseWriter.WriteHeader(http.StatusNotFound)
		} else {
			responseWriter.WriteHeader(http.StatusInternalServerError)
		}
		log.WithError(err).Error("unable to retrieve user")
		return
	}

	vocabService := dic.GetService[activitypub.VocabService]()
	followers, err := vocabService.GetFollowers(user)
	if err != nil {
		log.Error(err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Add("content-type", "application/activity+json")
	err = json.NewEncoder(responseWriter).Encode(followers)
	if err != nil {
		log.Error(err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func HandleOutbox(responseWriter http.ResponseWriter, request *http.Request) {
	log := dic.GetService[logger.Logger]()

	vars := mux.Vars(request)
	userService := dic.GetService[domain.UserService]()
	user, err := userService.GetFullUser(request.Context(), vars["username"])

	if err != nil {
		var twitterUserNotFound twitter.UsernameNotFoundError
		if errors.As(err, &twitterUserNotFound) || errors.Is(err, domain.ErrUserDoesNotExist) {
			responseWriter.WriteHeader(http.StatusNotFound)
		} else {
			responseWriter.WriteHeader(http.StatusInternalServerError)
		}
		log.WithError(err).Error("unable to retrieve user")
		return
	}

	vocabService := dic.GetService[activitypub.VocabService]()
	outbox, err := vocabService.GetOutbox(user)
	if err != nil {
		log.Error(err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Add("content-type", "application/activity+json")
	err = json.NewEncoder(responseWriter).Encode(outbox)
	if err != nil {
		log.Error(err)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func HandleInbox(responseWriter http.ResponseWriter, request *http.Request) {
	log := dic.GetService[logger.Logger]()

	if !auth.IsRequestSigned(request) {
		responseWriter.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(request)
	twitterClient := dic.GetService[twitter.TwitterClient]()
	_, err := twitterClient.GetUser(request.Context(), vars["username"])

	if err != nil {
		log.WithError(err).Error("unable to retrieve user")
		var twitterUserNotFound twitter.UsernameNotFoundError
		if errors.As(err, &twitterUserNotFound) {
			responseWriter.WriteHeader(http.StatusNotFound)
			return
		}
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	inboxService := dic.GetService[domain.InboxService]()
	var jsonMap map[string]interface{}
	jsonResolver, err := streams.NewJSONResolver(
		inboxService.Follow,
		inboxService.UnFollow,
	)
	if err != nil {
		log.WithError(err).Error("unable to create activity streams json resolver")
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewDecoder(request.Body).Decode(&jsonMap)
	if err != nil {
		log.WithError(err).Error("unable to decode activity streams json")
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}
	err = jsonResolver.Resolve(request.Context(), jsonMap)
	if err != nil {
		logEntry := log.WithField("activity", jsonMap["type"]).WithError(err)
		var notAllowedErr *domain.ActorNotAllowedError
		var notAllowedUndo *domain.UnsuportedUndoObjectError

		switch {
		case errors.As(err, &notAllowedErr):
			logEntry = logEntry.WithField("actor", notAllowedErr.Actor())
			responseWriter.WriteHeader(http.StatusForbidden)
		case errors.Is(err, streams.ErrNoCallbackMatch):
			fallthrough
		case errors.Is(err, domain.ErrFollowMismatchDomain):
			fallthrough
		case errors.As(err, &notAllowedUndo):
			fallthrough
		case errors.Is(err, domain.ErrUserDoesNotExist):
			responseWriter.WriteHeader(http.StatusBadRequest)
		default:
			responseWriter.WriteHeader(http.StatusInternalServerError)
		}

		logEntry.Error("unable to handle activity")
		return
	}

	responseWriter.WriteHeader(http.StatusAccepted)
}
