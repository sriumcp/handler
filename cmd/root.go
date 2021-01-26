package cmd

import (
	"fmt"
	"os"

	"github.com/iter8-tools/handler/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// config
var cfgFile string

// log
var log *logrus.Logger

// package constants
const version string = "v0.1.0-pre"

// package variables used for holding flag values
var action string
var filePath string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "handler",
	Short: "Handler performs start up and completion tasks within iter8 experiments",
	Long: `
iter8 launches jobs at the start and completition of an experiment, and executes the handler logic within the job's containers. 

For example, while starting an experiment, iter8 may populate the (bare) versionInfo section within the experiment resource object. While finishing an experiment, iter8 may rollforward or rollback versions, and may do so by interacting with Git repositories. This handler program provides the default implementation of a task runner which can execute these tasks during experiment start/finish.`,
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.handler.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".handler" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".handler")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	viper.AutomaticEnv() // read in environment variables that match

	ll, err := logrus.ParseLevel(viper.GetString("log_level"))
	if err == nil {
		utils.SetLogLevel(ll)
	}
	log = utils.GetLogger()
}
