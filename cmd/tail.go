// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"log"

	_ "github.com/minus5/svckit/dcy/lazy"

	"github.com/minus5/pitwall/monit"
	"github.com/minus5/svckit/dcy"
	"github.com/spf13/cobra"
)

var tailCmd = &cobra.Command{
	Use:   "tail",
	Short: "tail logs in datacenter <dc> for <service>",
	Long: `Tail logs in datacenter <dc> for <service>.
  If services is missing it will list all available services in <dc>.

  Examples:
    monit tail haproxy
    monit tail --dc pg1 haproxy
    monit tail backend_api -i request_logger -t url,method
    monit tail backend_api -i request_logger -a duration,status,code,lib
    monit tail backend_api -a listic -e request_logger.go:30`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("tail called")
		if len(args) > 1 {
			cmd.Usage()
			return
		}
		service := ""
		if len(args) == 1 {
			service = args[0]
		}

		if err := dcy.ConnectTo("192.168.10.203:8500"); err != nil {
			log.Fatal(err)
		}
		addr, err := dcy.Service("nsq-notifier")
		if err != nil {
			log.Fatal(err)
		}
		monit.Tail(addr.String(), service)
	},
}

func init() {
	rootCmd.AddCommand(tailCmd)
	//monitCmd.AddCommand(tailCmd)

	tailCmd.Flags().StringVarP(&dc, "dc", "d", "", "datacenter to find service")
	tailCmd.MarkFlagRequired("dc")
	// tailCmd.Flags().StringVarP(&service, "service", "s", "", "service to tail logs")
	// tailCmd.MarkFlagRequired("service")
}
