package domainmodels

import (
	"crypto"
	"net/url"
	"time"

	crypt "github.com/estrys/estrys/internal/crypto"
	"github.com/estrys/estrys/internal/dic"
)

type UserMetrics struct {
	Following, Followers, Tweets uint64
}

type User struct {
	Name            string
	Username        string
	Description     string
	ProfileImageURL *url.URL
	CreatedAt       time.Time
	Metrics         UserMetrics
	PublicKey       crypto.PublicKey
}

func (u User) PublicKeyPem() string {
	return dic.GetService[crypt.KeyManager]().FormatPubKey(u.PublicKey)
}
