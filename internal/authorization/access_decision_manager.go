package authorization

import (
	"github.com/estrys/estrys/internal/authorization/attributes"
	"github.com/estrys/estrys/internal/authorization/voter"
)

type AuthorizationChecker interface {
	IsGranted(any, attributes.Attribute) bool
}

type voterAuthorizationChecker struct {
	voters []voter.Voter
}

func NewVoterAuthorizationChecker(voters []voter.Voter) *voterAuthorizationChecker {
	return &voterAuthorizationChecker{voters: voters}
}

func (v *voterAuthorizationChecker) IsGranted(subject any, attr attributes.Attribute) bool {
	for _, v := range v.voters {
		if !v.Supports(subject) {
			continue
		}
		if v.Vote(subject, attr) == voter.AccessGranted {
			return true
		}
	}
	return false
}
