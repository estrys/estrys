package twitter

import (
	"fmt"
	"net/http"
)

type Authorizer struct {
	Token string
}

func (a Authorizer) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}
