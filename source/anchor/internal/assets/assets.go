package assets

import "embed"

//go:embed method.md
var Method string

//go:embed attributes.txt
var Attributes string

//go:embed beacon.yml
var Workflow string

//go:embed packs
var Packs embed.FS
