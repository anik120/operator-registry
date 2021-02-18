package declarative

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/operator-framework/operator-registry/pkg/registry"
)

const (
	manifestDir              = "./testdata/bundle/manifests"
	tmpTestDirPrefix         = "tmp_test_"
	expectedConfigsDirectory = "./testdata/expected-configs"
)

func TestPrintConfigCreatesNewConfig(t *testing.T) {

	// create bundle from manifests
	bundle, err := testBundle()
	require.NoError(t, err)
	// create temporary dir for test
	tmpDir, err := ioutil.TempDir("./", tmpTestDirPrefix)
	defer os.RemoveAll(tmpDir)
	require.NoError(t, err)

	err = printConfig(bundle, tmpDir)
	require.NoError(t, err)

	expectedConfigName := bundle.Package + ".json"
	expectedConfig, err := ioutil.ReadFile(filepath.Join(expectedConfigsDirectory, expectedConfigName))
	require.NoError(t, err)
	actualConfig, err := ioutil.ReadFile(filepath.Join(tmpDir, expectedConfigName))
	require.NoError(t, err)
	require.Equal(t, expectedConfig, actualConfig)
}

func TestPrintConfigAddsToExistingConfig(t *testing.T) {

	// create bundle from manifests
	bundle, err := testBundle()
	require.NoError(t, err)

	// create temporary dir for test
	tmpDir, err := ioutil.TempDir("./", tmpTestDirPrefix)
	defer os.RemoveAll(tmpDir)
	require.NoError(t, err)

	// create package config with just the package info
	expectedConfigName := bundle.Package + ".json"
	f, err := os.Create(filepath.Join(tmpDir, expectedConfigName))
	require.NoError(t, err)
	configPackage, err := NewConfigPackage(bundle)
	require.NoError(t, err)
	pkgJSON, err := json.MarshalIndent(configPackage, " ", "   ")
	require.NoError(t, err)
	f.Write(pkgJSON)
	f.Write([]byte("\n"))

	err = printConfig(bundle, tmpDir)
	require.NoError(t, err)

	expectedConfig, err := ioutil.ReadFile(filepath.Join(expectedConfigsDirectory, expectedConfigName))
	require.NoError(t, err)
	actualConfig, err := ioutil.ReadFile(filepath.Join(tmpDir, expectedConfigName))
	require.NoError(t, err)
	require.Equal(t, expectedConfig, actualConfig)
}

func testBundle() (*registry.Bundle, error) {
	bundle := registry.NewBundle("test", &registry.Annotations{
		PackageName:        "lib-bucket-provisioner",
		Channels:           "alpha,beta",
		DefaultChannelName: "alpha",
	})

	// Read all files in manifests directory
	items, err := ioutil.ReadDir(manifestDir)
	if err != nil {
		return nil, fmt.Errorf("error reading manifest directory: %s", err)
	}

	// unmarshal objects into unstructured
	unstObjs := []*unstructured.Unstructured{}
	for _, item := range items {
		fileWithPath := filepath.Join(manifestDir, item.Name())
		data, err := ioutil.ReadFile(fileWithPath)
		if err != nil {
			return nil, fmt.Errorf("error reading manifest directory file:%s", err)
		}

		dec := k8syaml.NewYAMLOrJSONDecoder(strings.NewReader(string(data)), 30)
		k8sFile := &unstructured.Unstructured{}
		err = dec.Decode(k8sFile)
		if err != nil {
			return nil, fmt.Errorf("error marshalling manifest into unstructured %s:%s", k8sFile, err)
		}
		unstObjs = append(unstObjs, k8sFile)
	}

	// add unstructured objects to test bundle
	for _, object := range unstObjs {
		bundle.Add(object)
	}
	return bundle, nil
}
