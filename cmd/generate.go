package cmd

import (
	"devforge/internal/config"
	"devforge/internal/project"
	"fmt"
	"github.com/spf13/cobra"
)

func newGenerateCommand() *cobra.Command {
	var force bool
	c := &cobra.Command{
		Use:   "generate",
		Short: "Generate missing workspace files from devforge.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(workspace)
			if err != nil {
				return err
			}
			changed, err := project.Generate(workspace, cfg, force)
			if err != nil {
				return err
			}
			fmt.Printf("Generated %d files. Existing files were preserved unless --force was supplied.\n", changed)
			return nil
		},
	}
	c.Flags().BoolVar(&force, "force", false, "overwrite generated files (review changes first)")
	return c
}
