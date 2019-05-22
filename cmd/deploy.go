package cmd

import (
	_ "github.com/minus5/svckit/dcy/lazy"

	"github.com/minus5/pitwall/deploy"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy <service>",
	Short: "Deploys service to a deployment",
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
		deploy.Run(dep, service, path, registry, image, noGit, consul)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVarP(&dep, "dep", "d", "", "deployment to deploy to")
	deployCmd.MarkFlagRequired("dep")
}
