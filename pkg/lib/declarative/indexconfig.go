package declarative

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/operator-framework/operator-registry/pkg/containertools"
	"github.com/operator-framework/operator-registry/pkg/image"
	"github.com/operator-framework/operator-registry/pkg/image/containerdregistry"
	"github.com/operator-framework/operator-registry/pkg/image/execregistry"
	"github.com/operator-framework/operator-registry/pkg/lib/certs"
	libregistry "github.com/operator-framework/operator-registry/pkg/lib/registry"
	registry "github.com/operator-framework/operator-registry/pkg/registry"

	"github.com/sirupsen/logrus"
)

// ConfigEditor is an implementation of IndexConfig
type ConfigEditor struct {
	Logger *logrus.Entry
}

// AddToConfig persists the bundle in it's package config
func (c ConfigEditor) AddToConfig(request AddConfigRequest) error {

	var reg image.Registry
	var rerr error
	switch request.ContainerTool {
	case containertools.NoneTool:
		rootCAs, err := certs.RootCAs(request.CaFile)
		if err != nil {
			return fmt.Errorf("failed to get RootCAs: %v", err)
		}
		reg, rerr = containerdregistry.NewRegistry(containerdregistry.SkipTLS(request.SkipTLS), containerdregistry.WithRootCAs(rootCAs))
	case containertools.PodmanTool:
		fallthrough
	case containertools.DockerTool:
		reg, rerr = execregistry.NewRegistry(request.ContainerTool, c.Logger, containertools.SkipTLS(request.SkipTLS))
	}
	if rerr != nil {
		return rerr
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
