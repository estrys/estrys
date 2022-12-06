package voter

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-fed/activity/streams/vocab"
	"github.com/pkg/errors"

	"github.com/estrys/estrys/internal/authorization/attributes"
)

type activityVoter struct {
	allowedUsers []string
}

func NewActivityVoter(allowedUsers []string) *activityVoter {
	return &activityVoter{allowedUsers: allowedUsers}
}

func (f *activityVoter) Supports(a any) bool {
	_, ok := a.(vocab.ActivityStreamsActorProperty)
	return ok
}

func (f *activityVoter) Vote(a any, attr attributes.Attribute) decision {
	act, _ := a.(vocab.ActivityStreamsActorProperty)

	if attr == attributes.CanFollow {
		return f.canFollow(act)
	}

	return AccessDenied
}

func (f *activityVoter) canFollow(actor vocab.ActivityStreamsActorProperty) decision {
	if actor.Len() != 1 {
		return AccessDenied
	}

	username, err := f.urlToUsername(actor.Begin().GetIRI())
	if err != nil {
		return AccessDenied
	}

	for _, allowedUser := range f.allowedUsers {
		if strings.HasSuffix(*username, allowedUser) {
			return AccessGranted
		}
	}

	return AccessDenied
}

func (f *activityVoter) urlToUsername(ur *url.URL) (*string, error) {
	splittedPath := strings.Split(ur.Path, "/")
	if len(splittedPath) <= 1 {
		return nil, errors.New("unable to find an username in the path")
	}
	userName := splittedPath[len(splittedPath)-1]
	fediverseUser := fmt.Sprintf("@%s@%s", userName, ur.Hostname())
	return &fediverseUser, nil
}
