//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . DockerfileGenerator
package containertools

import (
	"fmt"
	"sort"

	"github.com/sirupsen/logrus"
)

const (
	defaultBinarySourceImage    = "quay.io/operator-framework/upstream-opm-builder"
	DefaultDbLocation           = "/database/index.db"
	DefaultConfigFolderLocation = "/configs/"
	DbLocationLabel             = "operators.operatorframework.io.index.database.v1"
	ConfigsLocationLabel        = "operators.operatorframework.io.configs.v1"
)

// DockerfileGenerator defines functions to generate index dockerfiles
type DockerfileGenerator interface {
	GenerateIndexDockerfile(string, map[string]string) string
}

// IndexDockerfileGenerator struct implementation of DockerfileGenerator interface
type IndexDockerfileGenerator struct {
	Logger *logrus.Entry
}

// NewDockerfileGenerator is a constructor that returns a DockerfileGenerator
func NewDockerfileGenerator(logger *logrus.Entry) DockerfileGenerator {
	return &IndexDockerfileGenerator{
		Logger: logger,
	}
}

// GenerateIndexDockerfile builds a string representation of a dockerfile to use when building
// an operator-registry index image
func (g *IndexDockerfileGenerator) GenerateIndexDockerfile(binarySourceImage string, itemsToAdd map[string]string) string {
	var dockerfile string

	if binarySourceImage == "" {
		binarySourceImage = defaultBinarySourceImage
	}

	g.Logger.Info("Generating dockerfile")

	// From
	dockerfile += fmt.Sprintf("FROM %s\n", binarySourceImage)

	// Labels
	dockerfile += fmt.Sprintf("LABEL %s=%s\n", DbLocationLabel, DefaultDbLocation)
	dockerfile += fmt.Sprintf("LABEL %s=%s\n", ConfigsLocationLabel, DefaultConfigFolderLocation)

	// Content
	sortedsrc := make([]string, 0)
	for src, _ := range itemsToAdd {
		sortedsrc = append(sortedsrc, src)
	}
	sort.Strings(sortedsrc)
	for _, src := range sortedsrc {
		dockerfile += fmt.Sprintf("ADD %s %s\n", src, itemsToAdd[src])
	}
	dockerfile += fmt.Sprintf("EXPOSE 50051\n")
	dockerfile += fmt.Sprintf("ENTRYPOINT [\"/bin/opm\"]\n")
	dockerfile += fmt.Sprintf("CMD [\"registry\", \"serve\", \"--database\", \"%s\"]\n", DefaultDbLocation)

	return dockerfile
}
