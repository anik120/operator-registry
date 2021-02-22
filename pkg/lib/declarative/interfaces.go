package declarative

import (
	"github.com/operator-framework/operator-registry/pkg/containertools"
	"github.com/sirupsen/logrus"
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

// InspectIndexRequest represents a request to inspect an index
// for it's content (i.e package configs)
type InspectIndexRequest struct {
	Image    string
	PullTool containertools.ContainerTool
	CaFile   string
	SkipTLS  bool
}

// IndexConfig represents the content packages of an index
// in a declarative way using config files
type IndexConfig interface {
	AddToConfig(AddConfigRequest) error
	InspectIndex(InspectIndexRequest) error
}

// NewIndexConfig returns an IndexConfig
func NewIndexConfig(logger *logrus.Entry) IndexConfig {
	return ConfigEditor{
		Logger: logger,
	}
}
