package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	path     string
	registry string
	dc       string
	noGit    bool
	service  string
)

//var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pitwall",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pit.yaml)")

	//rootCmd.PersistentFlags().StringP("dc", "d", "", "datacenter")
	//viper.BindPFlag("dc", rootCmd.PersistentFlags().Lookup("dc"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// if cfgFile != "" {
	// 	// Use config file from the flag.
	// 	viper.SetConfigFile(cfgFile)
	// } else {
	// 	// Find home directory.
	// 	home, err := homedir.Dir()
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		os.Exit(1)
	// 	}

	// 	// Search config in home directory with name ".pit" (without extension).
	// 	viper.AddConfigPath(home)
	// 	viper.SetConfigName(".pit")
	// }

	// viper.AutomaticEnv() // read in environment variables that match

	// // If a config file is found, read it in.
	// if err := viper.ReadInConfig(); err == nil {
	// 	//fmt.Println("Using config file:", viper.ConfigFileUsed())
	// }
}
