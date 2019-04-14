package cmd

import (
	"github.com/minus5/pitwall/deploy"
	_ "github.com/minus5/svckit/dcy/lazy"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy <service>",
	Short: "Deploys service to a datacenter",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 1 {
			cmd.Usage()
			return
		}
		service := ""
		if len(args) == 1 {
			service = args[0]
		}
		deploy.Run(dc, service, path, registry, image, noGit, consul, dc)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVarP(&dc, "dc", "d", "", "datacenter to deploy to")
	deployCmd.MarkFlagRequired("dc")

	deployCmd.Flags().StringVar(&image, "image", "", "deploy this image instead of selecting from registry")
	deployCmd.Flags().StringVar(&registry, "registry", "registry.dev.minus5.hr", "docker images registry url")
}
