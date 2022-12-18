package models

import "time"

type ReferenceType string

const (
	ReferenceTypeRetweet ReferenceType = "retweeted"
)

type Tweet struct {
	ID               string        `json:"id"`
	AuthorID         string        `json:"-"`
	ReferencedType   ReferenceType `json:"-"`
	Text             string        `json:"text"`
	Published        time.Time     `json:"published"`
	Sensitive        bool          `json:"sensitive"`
	ReferencedTweets []Tweet       `json:"-"`
}

func (t *Tweet) IsRetweet() bool {
	for _, refTweet := range t.ReferencedTweets {
		if refTweet.ReferencedType == ReferenceTypeRetweet {
			return true
		}
	}
	return false
}
