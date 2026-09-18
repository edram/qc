package data

import _ "embed"

// Regions contains the canonical administrative regions and provider-specific codes.
//
//go:embed regions.json
var Regions []byte
