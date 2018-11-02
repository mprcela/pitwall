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
	"fmt"
	"syscall"

	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/log"
	"github.com/spf13/cobra"
)

// nsqCmd represents the nsq command
var mongoCmd = &cobra.Command{
	Use:   "mongo",
	Short: "Mongo database root command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mongo called")
	},
}

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Starts mongo shell",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		s := &shell{}
		if err := s.read(userFolder, server); err != nil {
			log.Fatal(err)
		}
		if err := s.run(); err != nil {
			log.Fatal(err)
		}
	},
}

var (
	userFolder string
	server     int
)

func init() {
	mongoCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(mongoCmd)

	shellCmd.Flags().IntVarP(&server, "server", "s", -1, "server no to connect to: 0,1,2")
	shellCmd.Flags().StringVarP(&userFolder, "user", "u", "reader", "consul kv folder with user credentials")
}

type shell struct {
	connectionString string
	username         string
	password         string
}

// read database connection string and credentials from consul
// root key is mongo/shell
func (s *shell) read(userFolder string, server int) error {
	if err := dcy.ConnectTo(consul); err != nil {
		return err
	}
	csKey := "mongo/shell/connectionString"
	if server >= 0 {
		csKey = fmt.Sprintf("mongo/shell/connectionString%d", server)
	}
	cs, err := dcy.KV(csKey)
	if err != nil {
		return err
	}
	un, err := dcy.KV(fmt.Sprintf("mongo/shell/%s/username", userFolder))
	if err != nil {
		return err
	}
	pwd, err := dcy.KV(fmt.Sprintf("mongo/shell/%s/password", userFolder))
	if err != nil {
		return err
	}

	s.connectionString = string(cs)
	s.username = string(un)
	s.password = string(pwd)
	return nil
}

// run starts mongo shell
// It is run through terminal (zsh) to collect user settings in ~/.mongorc.js.
func (s *shell) run() error {
	fmt.Printf("Connecting to %s as %s \n", s.connectionString, s.username)
	shellCmd := fmt.Sprintf("mongo \"%s\" --ssl --authenticationDatabase admin --username %s --password \"%s\"",
		s.connectionString, s.username, s.password)
	return syscall.Exec("/usr/local/bin/zsh", []string{"", "-c", shellCmd}, []string{})
}
