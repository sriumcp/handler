package cmd

import (
	"github.com/iter8-tools/handler/experiment"
	"github.com/spf13/cobra"
)

// dryrunCmd represents the dryrun command
var dryrunCmd = &cobra.Command{
	Use:     "dryrun",
	Short:   "Verify if task lists are well-defined in the experiment",
	Long:    `dryrun examines the spec.strategy.handlers section of the experiment yaml and finds errors. Errors are indicated if task lists or tasks are not well-specified, or if the handler finds references to unknown libraries or unknown tasks-types within libraries.`,
	Example: "dryrun -e experiment.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		if exp, err := (&experiment.Builder{}).FromFile(filePath).Build(); err == nil {
			log.Trace(exp, err)
			exp.DryRun()
			return
		}
		log.Error("cannot build experiment")
	},
}

func init() {
	rootCmd.AddCommand(dryrunCmd)
	dryrunCmd.PersistentFlags().StringVarP(&filePath, "experiment", "e", "", "path to experiment yaml (required)")
	dryrunCmd.MarkPersistentFlagRequired("experiment")
}
