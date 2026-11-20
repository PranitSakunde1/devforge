package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devforge/internal/config"
	"devforge/internal/project"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a workspace using the interactive wizard",
		RunE: func(cmd *cobra.Command, args []string) error {
			var name, dir string
			var goStack, javaStack, pythonStack, redisStack bool
			name, dir = "my-platform", workspace
			goStack, javaStack, pythonStack, redisStack = true, true, true, true
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Workspace name").Value(&name).Validate(func(v string) error {
						if strings.TrimSpace(v) == "" {
							return fmt.Errorf("workspace name is required")
						}
						return nil
					}),
					huh.NewInput().Title("Workspace directory").Value(&dir),
				),
				huh.NewGroup(
					huh.NewConfirm().Title("Include Go").Value(&goStack),
					huh.NewConfirm().Title("Include Java 21").Value(&javaStack),
					huh.NewConfirm().Title("Include Python").Value(&pythonStack),
					huh.NewConfirm().Title("Include Redis").Value(&redisStack),
				),
			)
			if err := form.Run(); err != nil {
				return err
			}
			if !filepath.IsAbs(dir) {
				abs, err := filepath.Abs(dir)
				if err != nil {
					return err
				}
				dir = abs
			}
			if _, err := os.Stat(filepath.Join(dir, "devforge.yaml")); err == nil {
				return fmt.Errorf("devforge.yaml already exists in %s; refusing to overwrite", dir)
			}
			cfg := config.Default(name)
			cfg.Stacks.Go, cfg.Stacks.Java, cfg.Stacks.Python, cfg.Stacks.Redis = goStack, javaStack, pythonStack, redisStack
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			if err := config.Save(dir, cfg); err != nil {
				return err
			}
			if err := project.Generate(dir, cfg, false); err != nil {
				return err
			}
			fmt.Printf("\nWorkspace created: %s\n", dir)
			fmt.Println("Next: devforge doctor -C", dir, "then devforge up -C", dir)
			return nil
		},
	}
}
