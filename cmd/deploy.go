package cmd

import (
	"github.com/minus5/pitwall/deploy"

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
		deploy.Run(dc, service, path, registry, noGit)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringVar(&path, "path", "~/work/pit/infrastructure", "infastructure project path")
	deployCmd.Flags().StringVar(&registry, "registry", "registry.dev.minus5.hr", "docker images registry url")
	deployCmd.Flags().BoolVar(&noGit, "no-git", false, "don't pull/push to infrastructure repository")
	deployCmd.Flags().StringVarP(&dc, "dc", "d", "", "datacenter to deploy to")
	deployCmd.MarkFlagRequired("dc")
}
