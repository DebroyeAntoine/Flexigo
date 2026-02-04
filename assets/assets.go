package assets

import "embed"

// ConfigFS contient le fichier config.yaml
//
//go:embed config.yaml
var ConfigFS embed.FS
