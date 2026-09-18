package data

import _ "embed"

// RegionsJSON is a snapshot of QCC's region response.
//
//go:embed qcc_regions.json
var RegionsJSON []byte
