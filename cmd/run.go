package cmd

import (
	"errors"
	"os"

	"github.com/iter8-tools/handler/experiment"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// getExperimentNN gets the name and namespace of the experiment from environment variables.
// Returns error if unsuccessful.
func getExperimentNN() (name string, namespace string, err error) {
	name = viper.GetViper().GetString("experiment_name")
	namespace = viper.GetViper().GetString("experiment_namespace")
	if len(name) == 0 || len(namespace) == 0 {
		return name, namespace, errors.New("invalid experiment name/namespace")
	}
	return name, namespace, nil
}

// run is a helper function used in the definition of runCmd cobra command.
func run(cmd *cobra.Command, args []string) error {
	name, namespace, err := getExperimentNN()
	if err == nil {
		var restConf *rest.Config
		restConf, err = config.GetConfig()
		if err == nil {
			var restClient client.Client
			restClient, err = experiment.GetClient(restConf)
			if err == nil {
				var exp *experiment.Experiment
				if exp, err = (&experiment.Builder{}).FromCluster(name, namespace, restClient).Build(); err == nil {
					err = exp.Run(action)
				}
			}
		}
	}
	return err
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run an action",
	Long:  `Sequentially execute all tasks in the specified action; if any task run results in an error, exit immediately with error.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(cmd, args); err != nil {
			log.Error(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().StringVarP(&action, "action", "a", "", "name of the action")
	runCmd.MarkPersistentFlagRequired("action")
}
