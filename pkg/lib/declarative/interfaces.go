package declarative

import (
	"github.com/sirupsen/logrus"

	"github.com/operator-framework/operator-registry/pkg/containertools"
)

// AddConfigRequest represents a request for a bundle to be added
// to it's package config
type AddConfigRequest struct {
	Bundles       []string
	ConfigFolder  string
	ContainerTool containertools.ContainerTool
	CaFile        string
	SkipTLS       bool
}

// IndexConfig represents the content packages of an index
// in a declarative way using config files
type IndexConfig interface {
	AddToConfig(AddConfigRequest) error
}

// NewIndexConfig returns an IndexConfig
func NewIndexConfig(logger *logrus.Entry) IndexConfig {
	return ConfigEditor{
		Logger: logger,
	}
}
