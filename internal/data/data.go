package data

import _ "embed"

// Regions contains the canonical administrative regions and provider-specific codes.
//
//go:embed regions.json
var Regions []byte

// Industries contains the canonical industry tree and provider-specific codes.
//
//go:embed industries.json
var Industries []byte
