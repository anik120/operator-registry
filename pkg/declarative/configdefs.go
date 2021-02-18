package declarative

import (
	"encoding/json"
	"strings"

	"github.com/operator-framework/operator-registry/pkg/registry"
)

const (
	bundleSchema    = "olm.bundle"
	packageSchema   = "olm.package"
	propertyChannel = "olm.channel"
)

type icon struct {
	Base64data string `json:"base64data"`
	Mediatype  string `json:"mediatype"`
}

type configPackage struct {
	Schema         string   `json:"schema"`
	Name           string   `json:"name"`
	DefaultChannel string   `json:"defaultChannel"`
	Icon           icon     `json:"icon"`
	Channels       []string `json:"channels"`
	Description    string   `json:"description"`
}

type channelPropertyValue struct {
	Name     string `json:"name"`
	Replaces string `json:"replaces"`
}
type configBundle struct {
	Schema        string               `json:"schema"`
	Name          string               `json:"name"`
	Package       string               `json:"package"`
	Image         string               `json:"image"`
	Version       string               `json:"version"`
	Properties    []*registry.Property `json:"properties"`
	RelatedImages []string             `json:"relatedImages"`
}

func NewConfigBundle(bundle *registry.Bundle) (*configBundle, error) {
	csv, err := bundle.ClusterServiceVersion()
	if err != nil {
		return nil, err
	}
	relatedImages, err := csv.GetRelatedImages()
	if err != nil {
		return nil, err
	}
	simpleRelatedImages := make([]string, 0)
	for key, _ := range relatedImages {
		simpleRelatedImages = append(simpleRelatedImages, key)
	}
	bundleReplaces, err := bundle.Replaces()
	if err != nil {
		return &configBundle{}, err
	}
	properties := bundle.Properties
	for _, channel := range bundle.Channels {
		channelPropertyValue, err := json.Marshal(channelPropertyValue{Name: channel, Replaces: bundleReplaces})
		if err != nil {
			return &configBundle{}, err
		}
		properties = append(properties, &registry.Property{Type: propertyChannel, Value: channelPropertyValue})
	}
	bundleSkips, err := bundle.Skips()
	if err != nil {
		return &configBundle{}, err
	}
	bundleSkipsJSON, err := json.Marshal(strings.Join(bundleSkips, ","))
	if err != nil {
		return &configBundle{}, err
	}
	bundleSkipsRange, err := bundle.SkipRange()
	if err != nil {
		return &configBundle{}, err
	}
	bundleSkipsRangeJSON, err := json.Marshal(bundleSkipsRange)
	if err != nil {
		return &configBundle{}, err
	}
	properties = append(properties, &registry.Property{Type: "skips", Value: bundleSkipsJSON})
	properties = append(properties, &registry.Property{Type: "skipsRange", Value: bundleSkipsRangeJSON})

	bundleVersion, err := bundle.Version()
	if err != nil {
		return &configBundle{}, err
	}
	return &configBundle{
		Schema:        bundleSchema,
		Name:          csv.Name,
		Package:       bundle.Package,
		Image:         bundle.BundleImage,
		Version:       bundleVersion,
		Properties:    properties,
		RelatedImages: simpleRelatedImages,
	}, nil
}

func NewConfigPackage(bundle *registry.Bundle) (*configPackage, error) {
	return &configPackage{
		Schema:         packageSchema,
		Name:           bundle.Package,
		DefaultChannel: bundle.Annotations.DefaultChannelName,
		Channels:       strings.Split(bundle.Annotations.Channels, ","),
		//TODO: Get icon and description
	}, nil
}
