package handlers

import (
	"net/http"
	"strings"
	"text/template"

	"github.com/estrys/estrys/internal/activitypub/handlers/views"
	"github.com/estrys/estrys/internal/config"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/domain"
)

func HandleWebFinger(responseWriter http.ResponseWriter, request *http.Request) {
	p := request.URL.Query().Get("resource")
	s := strings.Split(p, ":")
	splittedUserAddress := strings.Split(s[1], "@")
	username := splittedUserAddress[0]
	instance := splittedUserAddress[1]
	conf := dic.GetService[config.Config]()

	if instance != conf.Domain.Host {
		responseWriter.WriteHeader(http.StatusForbidden)
		return
	}

	userService := dic.GetService[domain.UserService]()
	user, err := userService.GetFullUser(request.Context(), username)
	if err != nil {
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	}

	templateContent, _ := views.Views.ReadFile("well_known/webfinger.json.tmpl")
	t := template.Must(template.New("webfinger").Parse(string(templateContent)))
	responseWriter.Header().Add("content-type", "application/json")
	_ = t.Execute(responseWriter, map[string]interface{}{
		"domain": conf.Domain,
		"user":   user,
	})
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
