package duh

import (
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// IsInitTemplateSpec checks if the OpenAPI spec contains all 4 required endpoints
// from the duh init template:
// - /v1/users.create
// - /v1/users.get
// - /v1/users.list
// - /v1/users.update
//
// Additional endpoints beyond these 4 are allowed and don't affect the match.
// Returns true only if ALL 4 required paths exist.
func IsInitTemplateSpec(spec *v3.Document) bool {
	if spec == nil || spec.Paths == nil || spec.Paths.PathItems == nil {
		return false
	}

	requiredPaths := []string{
		"/v1/users.create",
		"/v1/users.get",
		"/v1/users.list",
		"/v1/users.update",
	}

	for _, requiredPath := range requiredPaths {
		found := false
		for pair := orderedmap.First(spec.Paths.PathItems); pair != nil; pair = pair.Next() {
			if pair.Key() == requiredPath {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
