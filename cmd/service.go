package cmd

import (
	"fmt"

	"github.com/minus5/pitwall/service"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(subCmd)
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Get service information",
	//Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO")
	},
}

var subCmd = &cobra.Command{
	Use:   "list",
	Short: "List services location",
	//Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		pattern := ""
		if len(args) > 0 {
			pattern = args[0]
		}
		f := service.NewFinder(consul)
		all := f.All(pattern)
		for _, name := range all.Names() {
			fmt.Printf("%s\n", name)
			for _, svc := range all[name] {
				fmt.Printf("\t%s\n", svc)
			}
		}
	},
}
