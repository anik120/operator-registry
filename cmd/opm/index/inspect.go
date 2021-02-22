package index

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/operator-framework/operator-registry/pkg/containertools"
	"github.com/operator-framework/operator-registry/pkg/lib/declarative"
)

var (
	// TODO: Provide option for unpacking the representations as yaml
	inspectLong = templates.LongDesc(`
		Inspect an index for it's declarative package representations.

		This command unpacks the package representations (in json) that is stored within the index.

	`)
)

func newIndexInspectCmd() *cobra.Command {
	indexCmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect an index for it's content(i.e operators shipped with the index)",
		Long:  inspectLong,

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			return nil
		},

		RunE: runIndexInspectCmdFunc,
	}

	indexCmd.Flags().Bool("debug", false, "enable debug logging")
	indexCmd.Flags().StringP("image", "i", "", "container image of index to inspect")
	if err := indexCmd.MarkFlagRequired("image"); err != nil {
		logrus.Panic("Failed to set required `image` flag for `index inspect`")
	}
	indexCmd.Flags().StringP("pull-tool", "p", "", "tool to pull container images. One of: [none, docker, podman]. Defaults to none.")

	if err := indexCmd.Flags().MarkHidden("debug"); err != nil {
		logrus.Panic(err.Error())
	}

	return indexCmd
}

func runIndexInspectCmdFunc(cmd *cobra.Command, args []string) error {

	image, err := cmd.Flags().GetString("image")
	if err != nil {
		return err
	}
	pullTool, err := cmd.Flags().GetString("pull-tool")
	if err != nil {
		return err
	}
	inspectRequest := declarative.InspectIndexRequest{
		Image:    image,
		PullTool: containertools.NewContainerTool(pullTool, containertools.NoneTool),
	}
	logger := logrus.WithFields(logrus.Fields{"image": image})
	return declarative.NewIndexConfig(logger).InspectIndex(inspectRequest)
}
