package handlers

import (
	"net/http"
	"strings"
	"text/template"

	"github.com/estrys/estrys/internal/activitypub/handlers/views"
	"github.com/estrys/estrys/internal/config"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/domain"
	"github.com/estrys/estrys/internal/errors"
)

func HandleWebFinger(responseWriter http.ResponseWriter, request *http.Request) error {
	resource := request.URL.Query().Get("resource")
	resourceSplit := strings.Split(resource, ":")
	if len(resourceSplit) != 2 {
		return errors.HandlerError{
			UserMessage: "resource parameter should start with 'acct:'",
			HTTPCode:    http.StatusBadRequest,
		}
	}
	splittedUserAddress := strings.Split(resourceSplit[1], "@")
	if len(splittedUserAddress) != 2 {
		return errors.HandlerError{
			UserMessage: "username invalid",
			HTTPCode:    http.StatusBadRequest,
		}
	}
	username := splittedUserAddress[0]
	instance := splittedUserAddress[1]
	conf := dic.GetService[config.Config]()

	if instance != conf.Domain.Host {
		return errors.HandlerError{
			UserMessage: "requested user is not on this instance",
			HTTPCode:    http.StatusNotFound,
		}
	}

	userService := dic.GetService[domain.UserService]()
	user, err := userService.GetFullUser(request.Context(), username)
	if err != nil {
		return errors.HandlerError{
			Cause:       err,
			UserMessage: "user not found",
			HTTPCode:    http.StatusNotFound,
		}
	}

	templateContent, _ := views.Views.ReadFile("well_known/webfinger.json.tmpl")
	t := template.Must(template.New("webfinger").Parse(string(templateContent)))
	responseWriter.Header().Add("content-type", "application/json")
	_ = t.Execute(responseWriter, map[string]interface{}{
		"domain": conf.Domain,
		"user":   user,
	})

	return nil
}

func HandleHostMeta(w http.ResponseWriter, r *http.Request) {
	templateContent, _ := views.Views.ReadFile("well_known/host-meta.xml.tmpl")
	t := template.Must(template.New("webfinger").Parse(string(templateContent)))
	c := dic.GetService[config.Config]()
	w.Header().Add("content-type", "application/xrd+xml; charset=utf-8")
	_ = t.Execute(w, map[string]string{
		"baseUrl": c.Domain.String(),
	})
}
