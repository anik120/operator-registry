package declarative

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/operator-framework/operator-registry/pkg/containertools"
	"github.com/operator-framework/operator-registry/pkg/image"
	"github.com/operator-framework/operator-registry/pkg/image/containerdregistry"
	"github.com/operator-framework/operator-registry/pkg/image/execregistry"
	"github.com/operator-framework/operator-registry/pkg/lib/certs"
	libregistry "github.com/operator-framework/operator-registry/pkg/lib/registry"
	registry "github.com/operator-framework/operator-registry/pkg/registry"

	dircopy "github.com/otiai10/copy"
	"github.com/sirupsen/logrus"
)

const (
	DefaultInspectDir     = "configs"
	imgUnpackTmpDirPrefix = "tmp_unpack_"
)

// ConfigEditor is an implementation of IndexConfig
type ConfigEditor struct {
	Logger *logrus.Entry
}

// AddToConfig persists the bundle in it's package config
func (c ConfigEditor) AddToConfig(request AddConfigRequest) error {

	reg, err := imageRegistry(request.ContainerTool, request.CaFile, request.SkipTLS, c.Logger)
	if err != nil {
		return err
	}
	defer func() {
		if err := reg.Destroy(); err != nil {
			c.Logger.WithError(err).Warn("error destroying local cache")
		}
	}()

	simpleRefs := make([]image.Reference, 0)
	for _, ref := range request.Bundles {
		simpleRefs = append(simpleRefs, image.SimpleReference(ref))
	}

	for _, ref := range simpleRefs {
		to, from, cleanup, err := libregistry.UnpackImage(context.TODO(), reg, ref)
		if err != nil {
			return err
		}
		defer cleanup()

		img, err := registry.NewImageInput(to, from)
		if err := printConfig(img.Bundle, request.ConfigFolder); err != nil {
			return err
		}
	}

	return nil
}

// InspectIndex unpacks the package configs stored in an index image
// to a local directory
func (c ConfigEditor) InspectIndex(request InspectIndexRequest) error {

	// Pull the fromIndex
	c.Logger.Infof("Pulling image %s to get metadata", request.Image)

	reg, err := imageRegistry(request.PullTool, request.CaFile, request.SkipTLS, c.Logger)
	if err != nil {
		return err
	}
	defer func() {
		if err := reg.Destroy(); err != nil {
			c.Logger.WithError(err).Warn("error destroying local cache")
		}
	}()

	imageRef := image.SimpleReference(request.Image)
	if err := reg.Pull(context.TODO(), imageRef); err != nil {
		return fmt.Errorf("Error pulling image from remote registry. Error: %s", err)
	}
	// Get the index image's ConfigsLocation Label to find this path
	labels, err := reg.Labels(context.TODO(), imageRef)
	if err != nil {
		return err
	}

	configsLocation, ok := labels[containertools.ConfigsLocationLabel]
	if !ok {
		return fmt.Errorf("Index image %s missing label %s", request.Image, containertools.ConfigsLocationLabel)
	}

	tmpDir, err := ioutil.TempDir("./", imgUnpackTmpDirPrefix)
	defer os.RemoveAll(tmpDir)
	if err != nil {
		return err
	}
	reg.Unpack(context.TODO(), imageRef, tmpDir)

	configDir := filepath.Join("./", DefaultInspectDir)
	if err := os.MkdirAll(configDir, 0777); err != nil {
		return err
	}
	if err := dircopy.Copy(filepath.Join(tmpDir, configsLocation), configDir, dircopy.Options{}); err != nil {
		return err
	}
	c.Logger.Infof("Unpacked image %s to directory %s", request.Image, DefaultInspectDir)
	return nil
}

func printConfig(bundle *registry.Bundle, configFolder string) error {

	configFile := configFolder + "/" + bundle.Package + ".json"
	configBundle, err := NewConfigBundle(bundle)
	if err != nil {
		return err
	}
	bundleJSON, err := json.MarshalIndent(configBundle, " ", "   ")
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	var f *os.File
	defer f.Close()
	if _, err = os.Stat(configFile); err == nil {
		f, err = os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		f, err = os.Create(configFile)
		if err != nil {
			return err
		}
		configPackage, err := NewConfigPackage(bundle)
		if err != nil {
			return err
		}
		pkgJSON, err := json.MarshalIndent(configPackage, " ", "   ")
		if err != nil {
			return err
		}
		f.Write(pkgJSON)
		f.Write([]byte("\n"))
	} else {
		return err
	}
	f.Write(bundleJSON)
	f.Write([]byte("\n"))
	return err
}

func imageRegistry(tool containertools.ContainerTool, caFile string, skipTLS bool, logger *logrus.Entry) (image.Registry, error) {

	var reg image.Registry
	var rerr error

	switch tool {
	case containertools.NoneTool:
		rootCAs, err := certs.RootCAs(caFile)
		if err != nil {
			return nil, fmt.Errorf("failed to get RootCAs: %v", err)
		}
		reg, rerr = containerdregistry.NewRegistry(containerdregistry.SkipTLS(skipTLS), containerdregistry.WithLog(logger), containerdregistry.WithRootCAs(rootCAs))
	case containertools.PodmanTool:
		fallthrough
	case containertools.DockerTool:
		reg, rerr = execregistry.NewRegistry(tool, logger, containertools.SkipTLS(skipTLS))
	}
	if rerr != nil {
		return nil, rerr
	}
	return reg, rerr
}
