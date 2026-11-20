package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

func newInstallCommand() *cobra.Command {
	var yes bool
	c := &cobra.Command{
		Use:   "install",
		Short: "Install Windows prerequisites using winget (with consent)",
		Long:  "Installs Docker Desktop and Git through winget. Runtime toolchains remain containerized. This command never elevates privileges or silently accepts package agreements.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runtime.GOOS != "windows" {
				return fmt.Errorf("install is supported only on Windows")
			}
			if _, err := exec.LookPath("winget"); err != nil {
				return fmt.Errorf("winget not found; install App Installer from Microsoft Store")
			}
			packages := []struct{ id, name string }{
				{"Git.Git", "Git"},
				{"Docker.DockerDesktop", "Docker Desktop"},
			}
			reader := bufio.NewReader(os.Stdin)
			for _, p := range packages {
				listed, listErr := exec.CommandContext(cmd.Context(), "winget", "list", "--id", p.id, "--exact", "--source", "winget").CombinedOutput()
				if listErr == nil && strings.Contains(strings.ToLower(string(listed)), strings.ToLower(p.id)) {
					fmt.Printf("[OK] %s already appears in winget's installed package list; skipping.\n", p.name)
					continue
				}
				if !yes {
					fmt.Printf("Install %s using winget? [y/N] ", p.name)
					answer, _ := reader.ReadString('\n')
					if !strings.EqualFold(strings.TrimSpace(answer), "y") {
						fmt.Println("Skipped", p.name)
						continue
					}
				}
				args := []string{"install", "--id", p.id, "--exact", "--source", "winget", "--accept-source-agreements", "--accept-package-agreements"}
				fmt.Printf("\nInstalling %s...\n", p.name)
				c := exec.CommandContext(cmd.Context(), "winget", args...)
				c.Stdout, c.Stderr, c.Stdin = os.Stdout, os.Stderr, os.Stdin
				if err := c.Run(); err != nil {
					fmt.Printf("[ERROR] %s installation failed: %v\n", p.name, err)
					continue
				}
				fmt.Printf("[OK] %s installed. Restart terminal / Docker Desktop if required.\n", p.name)
			}
			fmt.Println("\nRun `devforge doctor` after Docker Desktop has started and switched to Linux containers.")
			return nil
		},
	}
	c.Flags().BoolVar(&yes, "yes", false, "skip per-package prompts (still uses winget package agreement flags)")
	return c
}
