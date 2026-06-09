package docs

import _ "embed"

// redocJS is the vendored ReDoc standalone bundle, embedded so that both serve
// and export deliver the renderer without any network access (ADR-0008, BC-1).
//
//go:embed redoc.standalone.js
var redocJS []byte
