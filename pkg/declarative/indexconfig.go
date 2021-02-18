package declarative

import (
	"context"
	"encoding/json"
	"os"

	"github.com/operator-framework/operator-registry/pkg/image"
	libregistry "github.com/operator-framework/operator-registry/pkg/lib/registry"
	registry "github.com/operator-framework/operator-registry/pkg/registry"

	"github.com/sirupsen/logrus"
)

// ConfigEditor is an implementation of IndexConfig
type ConfigEditor struct {
	Logger      *logrus.Entry
	ImgRegistry image.Registry
}

// AddToConfig persists the bundle in it's package config
func (c ConfigEditor) AddToConfig(request AddConfigRequest) error {

	defer func() {
		if err := c.ImgRegistry.Destroy(); err != nil {
			c.Logger.WithError(err).Warn("error destroying local cache")
		}
	}()

	simpleRefs := make([]image.Reference, 0)
	for _, ref := range request.Bundles {
		simpleRefs = append(simpleRefs, image.SimpleReference(ref))
	}

	for _, ref := range simpleRefs {
		to, from, cleanup, err := libregistry.UnpackImage(context.TODO(), c.ImgRegistry, ref)
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
		writeJSONToFile(pkgJSON, f)
	} else {
		return err
	}
	writeJSONToFile(bundleJSON, f)
	return err
}

func writeJSONToFile(json []byte, file *os.File) {
	file.Write(json)
	file.Write([]byte("\n"))
}
