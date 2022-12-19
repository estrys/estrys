package models

import (
	"net/url"
	"strings"
	"time"
)

type ReferenceType string

const (
	ReferenceTypeRetweet ReferenceType = "retweeted"
)

type MediaType string

const (
	MediaTypePhoto MediaType = "photo"
)

type TweetMedia struct {
	Type          MediaType
	URL           *url.URL
	Width, Height int
}

type Tweet struct {
	ID               string
	AuthorID         string
	AuthorUsername   string
	ReferencedType   ReferenceType
	Text             string
	Published        time.Time
	Sensitive        bool
	ReferencedTweets []Tweet
	Medias           []TweetMedia
}

func (t *Tweet) IsAuthoredBy(username string) bool {
	return strings.EqualFold(t.AuthorUsername, username)
}

func (t *Tweet) Retweet() *Tweet {
	for _, refTweet := range t.ReferencedTweets {
		if refTweet.ReferencedType == ReferenceTypeRetweet {
			return &refTweet
		}
	}
	return nil
}
