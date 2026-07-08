package rules

import (
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

	"github.com/duh-rpc/duh-cli/internal/didyoumean"
)

// DidYouMeanCollisionRule enforces the switch-path uniqueness invariant on the
// x-duh-did-you-mean teaching extension: every entry must sit beside a post:
// operation, be a well-formed spec-relative path, must not equal a canonical
// route, and must be declared exactly once across the spec. It gives authors the
// same feedback 'duh generate' enforces, only earlier — from the same
// implementation in internal/didyoumean, so the two cannot drift.
type DidYouMeanCollisionRule struct{}

func NewDidYouMeanCollisionRule() *DidYouMeanCollisionRule {
	return &DidYouMeanCollisionRule{}
}

func (r *DidYouMeanCollisionRule) Name() string {
	return "DID_YOU_MEAN_COLLISION"
}

func (r *DidYouMeanCollisionRule) Validate(doc *v3.Document) []Violation {
	problems := didyoumean.Validate(doc)
	if len(problems) == 0 {
		return nil
	}
	violations := make([]Violation, 0, len(problems))
	for _, p := range problems {
		location := p.Path
		if location == "" {
			location = p.Entry
		}
		violations = append(violations, Violation{
			Suggestion: p.Suggestion,
			Message:    p.Message,
			Location:   location,
			RuleName:   r.Name(),
			Severity:   SeverityError,
		})
	}
	return violations
}
