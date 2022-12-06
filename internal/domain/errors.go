package domain

import "github.com/pkg/errors"

var ErrFollowMismatchDomain = errors.New("unable to follow an user outside this instance")
var ErrUserDoesNotExist = errors.New("user does not exist")
