package voter

import (
	"github.com/estrys/estrys/internal/authorization/attributes"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func Test_activityVoter_Vote_CanFollow(t *testing.T) {
	tests := []struct {
		name         string
		subject      any
		attr         attributes.Attribute
		allowedUsers []string
		want         decision
	}{
		{
			name: "access granted user match",
			subject: func() vocab.ActivityStreamsActorProperty {
				prop := streams.NewActivityStreamsActorProperty()
				url, _ := url.Parse("https://example.com/users/foobar")
				prop.AppendIRI(url)
				return prop
			}(),
			attr:         attributes.CanFollow,
			allowedUsers: []string{"@foobar@example.com"},
			want:         AccessGranted,
		},
		{
			name: "access granted instance match",
			subject: func() vocab.ActivityStreamsActorProperty {
				prop := streams.NewActivityStreamsActorProperty()
				url, _ := url.Parse("https://example.com/users/foobar")
				prop.AppendIRI(url)
				return prop
			}(),
			attr:         attributes.CanFollow,
			allowedUsers: []string{"@example.com"},
			want:         AccessGranted,
		},
		{
			name: "access denier no allowed users",
			subject: func() vocab.ActivityStreamsActorProperty {
				prop := streams.NewActivityStreamsActorProperty()
				url, _ := url.Parse("https://example.com/users/foobar")
				prop.AppendIRI(url)
				return prop
			}(),
			attr:         attributes.CanFollow,
			allowedUsers: []string{},
			want:         AccessDenied,
		},
		{
			name: "access denier multiples actors",
			subject: func() vocab.ActivityStreamsActorProperty {
				prop := streams.NewActivityStreamsActorProperty()
				url, _ := url.Parse("https://example.com/users/foobar")
				prop.AppendIRI(url)
				prop.AppendIRI(url)
				return prop
			}(),
			attr:         attributes.CanFollow,
			allowedUsers: []string{"@example.com"},
			want:         AccessDenied,
		},
		{
			name: "access denier invalid actor url",
			subject: func() vocab.ActivityStreamsActorProperty {
				prop := streams.NewActivityStreamsActorProperty()
				url, _ := url.Parse("mailto:example.com")
				prop.AppendIRI(url)
				return prop
			}(),
			attr:         attributes.CanFollow,
			allowedUsers: []string{},
			want:         AccessDenied,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewActivityVoter(tt.allowedUsers)
			require.True(t, f.Supports(tt.subject))
			if got := f.Vote(tt.subject, tt.attr); got != tt.want {
				t.Errorf("Vote() = %v, want %v", got, tt.want)
			}
		})
	}
}
