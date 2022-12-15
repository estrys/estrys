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
	internalerrors "github.com/estrys/estrys/internal/errors"
	"github.com/estrys/estrys/internal/twitter"
)

func HandleUser(responseWriter http.ResponseWriter, request *http.Request) error {
	vars := mux.Vars(request)
	// TODO Validate username input

	userService := dic.GetService[domain.UserService]()
	user, err := userService.GetFullUser(request.Context(), vars["username"])

	if err != nil {
		var twitterUserNotFound twitter.UsernameNotFoundError
		if errors.As(err, &twitterUserNotFound) || errors.Is(err, domain.ErrUserDoesNotExist) {
			return internalerrors.HandlerError{
				Cause:       err,
				UserMessage: "user not found",
				HTTPCode:    http.StatusNotFound,
			}
		}
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}

	vocabService := dic.GetService[activitypub.VocabService]()
	actor, err := vocabService.GetActor(user)
	if err != nil {
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}

	responseWriter.Header().Add("content-type", "application/activity+json")
	err = json.NewEncoder(responseWriter).Encode(actor)
	if err != nil {
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}
	return nil
}

func HandleFollowing(responseWriter http.ResponseWriter, request *http.Request) error {
	vars := mux.Vars(request)
	userService := dic.GetService[domain.UserService]()
	user, err := userService.GetFullUser(request.Context(), vars["username"])

	if err != nil {
		var twitterUserNotFound twitter.UsernameNotFoundError
		if errors.As(err, &twitterUserNotFound) || errors.Is(err, domain.ErrUserDoesNotExist) {
			return internalerrors.HandlerError{
				Cause:       err,
				UserMessage: "user not found",
				HTTPCode:    http.StatusNotFound,
			}
		}
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}

	vocabService := dic.GetService[activitypub.VocabService]()
	following, err := vocabService.GetFollowing(user)
	if err != nil {
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}

	responseWriter.Header().Add("content-type", "application/activity+json")
	err = json.NewEncoder(responseWriter).Encode(following)
	if err != nil {
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}
	return nil
}

func HandleFollowers(responseWriter http.ResponseWriter, request *http.Request) error {
	vars := mux.Vars(request)
	userService := dic.GetService[domain.UserService]()
	user, err := userService.GetFullUser(request.Context(), vars["username"])

	if err != nil {
		var twitterUserNotFound twitter.UsernameNotFoundError
		if errors.As(err, &twitterUserNotFound) || errors.Is(err, domain.ErrUserDoesNotExist) {
			return internalerrors.HandlerError{
				Cause:       err,
				UserMessage: "user not found",
				HTTPCode:    http.StatusNotFound,
			}
		}
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}

	vocabService := dic.GetService[activitypub.VocabService]()
	followers, err := vocabService.GetFollowers(user)
	if err != nil {
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}

	responseWriter.Header().Add("content-type", "application/activity+json")
	err = json.NewEncoder(responseWriter).Encode(followers)
	if err != nil {
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}

	return nil
}

func HandleOutbox(responseWriter http.ResponseWriter, request *http.Request) error {
	vars := mux.Vars(request)
	userService := dic.GetService[domain.UserService]()
	user, err := userService.GetFullUser(request.Context(), vars["username"])

	if err != nil {
		var twitterUserNotFound twitter.UsernameNotFoundError
		if errors.As(err, &twitterUserNotFound) || errors.Is(err, domain.ErrUserDoesNotExist) {
			return internalerrors.HandlerError{
				Cause:       err,
				UserMessage: "user not found",
				HTTPCode:    http.StatusNotFound,
			}
		}
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}

	vocabService := dic.GetService[activitypub.VocabService]()
	outbox, err := vocabService.GetOutbox(user)
	if err != nil {
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}

	responseWriter.Header().Add("content-type", "application/activity+json")
	err = json.NewEncoder(responseWriter).Encode(outbox)
	if err != nil {
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}
	return nil
}

func HandleInbox(responseWriter http.ResponseWriter, request *http.Request) error {
	if !auth.IsRequestSigned(request) {
		return internalerrors.HandlerError{
			UserMessage: "request signature failed",
			HTTPCode:    http.StatusForbidden,
		}
	}

	vars := mux.Vars(request)
	twitterClient := dic.GetService[twitter.TwitterClient]()
	_, err := twitterClient.GetUser(request.Context(), vars["username"])

	if err != nil {
		var twitterUserNotFound twitter.UsernameNotFoundError
		if errors.As(err, &twitterUserNotFound) {
			return internalerrors.HandlerError{
				Cause:       err,
				UserMessage: "user not found",
				HTTPCode:    http.StatusNotFound,
			}
		}
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}

	inboxService := dic.GetService[domain.InboxService]()
	var jsonMap map[string]interface{}
	jsonResolver, err := streams.NewJSONResolver(
		inboxService.Follow,
		inboxService.UnFollow,
	)
	if err != nil {
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}
	err = json.NewDecoder(request.Body).Decode(&jsonMap)
	if err != nil {
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}
	err = jsonResolver.Resolve(request.Context(), jsonMap)
	if err != nil {
		var notAllowedErr *domain.ActorNotAllowedError
		var notAllowedUndo *domain.UnsuportedUndoObjectError
		switch {
		case errors.As(err, &notAllowedErr):
			return internalerrors.HandlerError{
				Cause:       err,
				UserMessage: "not allowed to follow this user",
				HTTPCode:    http.StatusForbidden,
			}
		case errors.Is(err, streams.ErrNoCallbackMatch):
			return internalerrors.HandlerError{
				Cause:       err,
				UserMessage: "unsupported activity",
				HTTPCode:    http.StatusBadRequest,
			}
		case errors.Is(err, domain.ErrFollowMismatchDomain):
			return internalerrors.HandlerError{
				Cause:       err,
				UserMessage: "cannot follow a user that is not on this instance",
				HTTPCode:    http.StatusBadRequest,
			}
		case errors.As(err, &notAllowedUndo):
			return internalerrors.HandlerError{
				Cause:       err,
				UserMessage: "can only undo Follow activities",
				HTTPCode:    http.StatusBadRequest,
			}
		case errors.Is(err, domain.ErrUserDoesNotExist):
			return internalerrors.HandlerError{
				Cause:       err,
				UserMessage: "user not found",
				HTTPCode:    http.StatusBadRequest,
			}
		}
		return internalerrors.HandlerError{
			Cause:    err,
			HTTPCode: http.StatusInternalServerError,
		}
	}
	responseWriter.WriteHeader(http.StatusAccepted)
	return nil
}
