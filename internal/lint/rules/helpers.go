package rules

import "strings"

func isPaginatedEndpoint(path string) bool {
	return strings.HasSuffix(path, ".list") ||
		strings.HasSuffix(path, ".search") ||
		strings.HasSuffix(path, ".query")
}

// isUnspecifiedSentinel mirrors fieldmap.isUnspecifiedSentinel: an enum variant is a
// sentinel when its value ends in UNSPECIFIED. The two definitions are intentionally
// identical; the lint↔generate parity invariant depends on them staying in lock-step
// (a spec passes the lint rule iff it passes fieldmap.assertUnspecifiedFirst).
func isUnspecifiedSentinel(value string) bool {
	return strings.HasSuffix(value, "UNSPECIFIED")
}

// isIntegerEnum reports whether a schema declares type: integer, the only OpenAPI
// shape DUH projects to a closed Protobuf enum (and thus the only shape where an
// *_UNSPECIFIED sentinel owns number 0).
func isIntegerEnum(types []string) bool {
	return len(types) > 0 && types[0] == "integer"
}

// isStringEnum reports whether a schema declares type: string, which DUH projects to
// an open Protobuf string field rather than a closed enum.
func isStringEnum(types []string) bool {
	return len(types) > 0 && types[0] == "string"
}
