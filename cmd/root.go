package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var workspace string

var rootCmd = &cobra.Command{
	Use:   "devforge",
	Short: "Windows-first Docker development environment manager",
	Long:  "DevForge provisions and operates reproducible Go, Java 21, Python and Redis development workspaces.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "devforge:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&workspace, "dir", "C", ".", "workspace directory")
	rootCmd.AddCommand(newInitCommand(), newDoctorCommand(), newInstallCommand(),
		newGenerateCommand(), newUpCommand(), newDownCommand(), newWatchCommand(),
		newLogsCommand(), newTestCommand(), newSecurityCommand(), newReportCommand(),
		newCleanCommand(), newVersionCommand())
}
