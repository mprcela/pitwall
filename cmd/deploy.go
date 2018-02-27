package cmd

import (
	"fmt"

	_ "github.com/minus5/svckit/dcy/lazy"

	"github.com/minus5/pitwall/deploy"
	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/log"
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
		deploy.Run(dc, service, path, registry, image, noGit, getServiceAddressByTag("http", "nomad"))
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVarP(&dc, "dc", "d", "", "datacenter to deploy to")
	deployCmd.MarkFlagRequired("dc")
}

func getServiceAddressByTag(tag, name string) string {
	if err := dcy.ConnectTo(consul); err != nil {
		log.Fatal(err)
	}
	addr, err := dcy.ServiceInDcByTag(tag, name, dc)
	if err == nil {
		return addr.String()
	}
	log.Fatal(fmt.Errorf("service %s with tag %s not found in consul %s", name, tag, consul))
	return ""
}
