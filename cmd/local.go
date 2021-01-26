package cmd

import (
	"os"

	"github.com/iter8-tools/handler/experiment"
	"github.com/spf13/cobra"
)

// localCmd represents the local command
var localCmd = &cobra.Command{
	Use:     "local",
	Short:   "Run tasks for an action from an experiment file.",
	Long:    `The normal run command fetches the experiment object from the cluster. This local subcommand uses a locally specific experiment file to run tasks within an action. The local subcommand is intended for testing purposes.`,
	Example: "local -e experiment.yaml -t start",
	Run: func(cmd *cobra.Command, args []string) {
		if exp, err := (&experiment.Builder{}).FromFile(filePath).Build(); err == nil {
			if err = exp.Run(action); err == nil {
				return
			}
		}
		log.Error("cannot run local")
		os.Exit(1)
	},
}

func init() {
	runCmd.AddCommand(localCmd)
	localCmd.PersistentFlags().StringVarP(&filePath, "experiment", "e", "", "path to experiment yaml file (required)")
	localCmd.MarkPersistentFlagRequired("experiment")
	localCmd.PersistentFlags().StringVarP(&action, "action", "a", "", "name of the action to run (required)")
	localCmd.MarkPersistentFlagRequired("action")
}
