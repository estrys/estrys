package voter

import "github.com/estrys/estrys/internal/authorization/attributes"

type decision int

const (
	AccessDenied decision = iota
	AccessGranted
)

type Voter interface {
	Supports(any) bool
	Vote(any, attributes.Attribute) decision
}
