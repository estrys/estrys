package domain

import (
	"fmt"

	"github.com/pkg/errors"
)

var ErrFollowMismatchDomain = errors.New("unable to follow an user outside this instance")
var ErrUserDoesNotExist = errors.New("user does not exist")

type TwitterUserDoesNotExistError struct {
	Username string
}

func (e *TwitterUserDoesNotExistError) Error() string {
	return fmt.Sprintf("@%s does not exist on twitter", e.Username)
}
