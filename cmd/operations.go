package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"devforge/internal/config"
	"github.com/spf13/cobra"
)

func compose(ctx context.Context, dir string, args ...string) error {
	file := filepath.Join(dir, "compose.yaml")
	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("compose.yaml not found in %s; run `devforge init`", dir)
	}
	argv := append([]string{"compose", "-f", file}, args...)
	c := exec.CommandContext(ctx, "docker", argv...)
	c.Dir, c.Stdout, c.Stderr, c.Stdin = dir, os.Stdout, os.Stderr, os.Stdin
	return c.Run()
}

func newUpCommand() *cobra.Command {
	return &cobra.Command{Use: "up", Short: "Build and start selected development services", RunE: func(c *cobra.Command, a []string) error {
		return compose(c.Context(), workspace, "up", "-d", "--build")
	}}
}
func newDownCommand() *cobra.Command {
	return &cobra.Command{Use: "down", Short: "Stop services and preserve named volumes", RunE: func(c *cobra.Command, a []string) error {
		return compose(c.Context(), workspace, "down", "--remove-orphans")
	}}
}
func newWatchCommand() *cobra.Command {
	return &cobra.Command{Use: "watch", Short: "Build, start and watch selected services for source changes", RunE: func(c *cobra.Command, a []string) error {
		return compose(c.Context(), workspace, "up", "--build", "--watch")
	}}
}
func newLogsCommand() *cobra.Command {
	return &cobra.Command{Use: "logs", Short: "Follow workspace container logs", RunE: func(c *cobra.Command, a []string) error {
		return compose(c.Context(), workspace, "logs", "-f", "--tail", "150")
	}}
}
func newTestCommand() *cobra.Command {
	return &cobra.Command{Use: "test", Short: "Run all enabled unit test suites and retain reports", RunE: func(c *cobra.Command, a []string) error {
		cfg, err := config.Load(workspace)
		if err != nil {
			return fmt.Errorf("load workspace configuration: %w", err)
		}
		suites := make([]string, 0, 3)
		if cfg.Stacks.Go {
			suites = append(suites, "go-test")
		}
		if cfg.Stacks.Java {
			suites = append(suites, "java-test")
		}
		if cfg.Stacks.Python {
			suites = append(suites, "python-test")
		}
		if len(suites) == 0 {
			return fmt.Errorf("no language test suites are enabled in devforge.yaml")
		}
		var failures []string
		for _, service := range suites {
			fmt.Printf("\n===== %s =====\n", service)
			ctx, cancel := context.WithTimeout(c.Context(), 20*time.Minute)
			err := compose(ctx, workspace, "run", "--rm", service)
			cancel()
			if err != nil {
				failures = append(failures, service+": "+err.Error())
			}
		}
		if err := generateReport(workspace); err != nil {
			failures = append(failures, "report generation: "+err.Error())
		}
		if len(failures) > 0 {
			return fmt.Errorf("test run completed with errors: %s", strings.Join(failures, "; "))
		}
		return nil
	}}
}
func newSecurityCommand() *cobra.Command {
	var image string
	c := &cobra.Command{Use: "security", Short: "Run Trivy filesystem scan and optionally scan an image", RunE: func(cmd *cobra.Command, args []string) error {
		if err := os.MkdirAll(filepath.Join(workspace, "reports", "security"), 0755); err != nil {
			return err
		}
		base := []string{"run", "--rm", "trivy"}
		run := func(args ...string) error {
			return compose(cmd.Context(), workspace, append(append([]string{}, base...), args...)...)
		}
		if err := run("fs", "--scanners", "vuln,misconfig,secret", "--format", "json", "--output", "/reports/trivy-fs.json", "/workspace"); err != nil {
			return err
		}
		if err := run("fs", "--scanners", "vuln,misconfig,secret", "--format", "sarif", "--output", "/reports/trivy-fs.sarif", "/workspace"); err != nil {
			return err
		}
		if image != "" {
			if err := run("image", "--format", "json", "--output", "/reports/trivy-image.json", image); err != nil {
				return err
			}
			if err := run("image", "--format", "sarif", "--output", "/reports/trivy-image.sarif", image); err != nil {
				return err
			}
		}
		return generateReport(workspace)
	}}
	c.Flags().StringVar(&image, "image", "", "also scan a local/remote container image")
	return c
}
func newReportCommand() *cobra.Command {
	return &cobra.Command{Use: "report", Short: "Generate consolidated HTML test and security report", RunE: func(cmd *cobra.Command, args []string) error {
		return generateReport(workspace)
	}}
}
func newCleanCommand() *cobra.Command {
	var volumes bool
	c := &cobra.Command{Use: "clean", Short: "Remove Compose containers; preserve data volumes by default", RunE: func(cmd *cobra.Command, args []string) error {
		argv := []string{"down", "--remove-orphans"}
		if volumes {
			fmt.Fprintln(os.Stderr, "WARNING: this permanently deletes named volumes, including Redis data.")
			fmt.Fprint(os.Stderr, "Type DELETE to continue: ")
			answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			if strings.TrimSpace(answer) != "DELETE" {
				return fmt.Errorf("cancelled")
			}
			argv = append(argv, "--volumes")
		}
		return compose(cmd.Context(), workspace, argv...)
	}}
	c.Flags().BoolVar(&volumes, "volumes", false, "also delete persistent volumes (destructive; requires typed confirmation)")
	return c
}
func newVersionCommand() *cobra.Command {
	return &cobra.Command{Use: "version", Short: "Print DevForge version", Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("devforge", version)
	}}
}
