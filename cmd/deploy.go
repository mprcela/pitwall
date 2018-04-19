package cmd

import (
	"fmt"

	_ "github.com/minus5/svckit/dcy/lazy"
	"github.com/minus5/svckit/env"

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

		nomadName := findNomadForService(service)
		deploy.Run(dc, service, path, registry, image, noGit, getServiceAddressByTag("http", nomadName))
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVarP(&dc, "dc", "d", "", "datacenter to deploy to")
	deployCmd.MarkFlagRequired("dc")
}

func findNomadForService(service string) string {
	c, err := deploy.NewDcConfig(env.ExpandPath(path), dc)
	if err != nil {
		log.Fatal(err)
	}
	svc := c.Find(service)
	if svc == nil {
		log.Fatal(fmt.Errorf("service %s not found", service))
	}
	name := "nomad"
	if svc.DcRegion == "s2" {
		name = "nomad-s2"
	}
	return name
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
