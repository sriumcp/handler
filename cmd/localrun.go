package cmd

import (
	"github.com/iter8-tools/handler/experiment"
	"github.com/spf13/cobra"
)

// localrunCmd represents the dryrun command
var localrunCmd = &cobra.Command{
	Use:     "localrun",
	Short:   "run task or action locally",
	Long:    `locally run a task or action; not all tasks support local runs; an action can be run locally if all its tasks support local runs.`,
	Example: "localrun -e experiment.yaml -a start -t 5",
	Run: func(cmd *cobra.Command, args []string) {
		if exp, err := (&experiment.Builder{}).FromFile(filePath).Build(); err == nil {
			log.Trace(exp, err)
			err = exp.LocalRun(action, task)
			if err == nil {
				return
			}
			log.Error(err)
		} else {
			log.Error("cannot build experiment", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(localrunCmd)
	localrunCmd.PersistentFlags().StringVarP(&filePath, "experiment", "e", "", "path to experiment yaml (required)")
	localrunCmd.MarkPersistentFlagRequired("experiment")
	localrunCmd.PersistentFlags().StringVarP(&action, "action", "a", "", "name of the action")
	localrunCmd.MarkPersistentFlagRequired("action")
	localrunCmd.PersistentFlags().IntVarP(&task, "task", "t", -1, "index of the task")
}
