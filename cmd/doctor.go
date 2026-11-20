package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var composeVersionPattern = regexp.MustCompile(`v?(\d+)\.(\d+)\.(\d+)`)

func newDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose Windows and Docker prerequisites",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Host: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			if runtime.GOOS != "windows" {
				fmt.Println("[INFO] DevForge CLI can run on non-Windows hosts; the winget installer is Windows-only.")
			}
			dockerPath, err := exec.LookPath("docker")
			if err != nil {
				return fmt.Errorf("Docker CLI not found; install Docker Desktop with `devforge install`")
			}
			out, err := runCheck(cmd.Context(), dockerPath, "--version")
			if err != nil {
				return fmt.Errorf("Docker CLI check failed: %w", err)
			}
			fmt.Println("[OK] Docker:", strings.TrimSpace(string(out)))

			out, err = runCheck(cmd.Context(), dockerPath, "compose", "version", "--short")
			if err != nil {
				return fmt.Errorf("Docker Compose v2 plugin is missing: %w", err)
			}
			version := strings.TrimSpace(string(out))
			if !atLeastCompose(version, 2, 23, 0) {
				return fmt.Errorf("Docker Compose %s detected; DevForge requires Compose v2.23.0 or newer for its configured Watch actions", version)
			}
			fmt.Println("[OK] Docker Compose:", version)

			out, err = runCheck(cmd.Context(), dockerPath, "info", "--format", "{{.OSType}}")
			if err != nil {
				return fmt.Errorf("Docker daemon is not ready; start Docker Desktop: %w", err)
			}
			engine := strings.TrimSpace(string(out))
			if engine != "linux" {
				return fmt.Errorf("Docker engine reports %q; switch Docker Desktop to Linux containers", engine)
			}
			fmt.Println("[OK] Docker engine: Linux containers")
			return nil
		},
	}
}

func runCheck(parent context.Context, bin string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, bin, args...).CombinedOutput()
}

func atLeastCompose(version string, major, minor, patch int) bool {
	m := composeVersionPattern.FindStringSubmatch(version)
	if len(m) != 4 {
		return false
	}
	a, _ := strconv.Atoi(m[1])
	b, _ := strconv.Atoi(m[2])
	c, _ := strconv.Atoi(m[3])
	if a != major {
		return a > major
	}
	if b != minor {
		return b > minor
	}
	return c >= patch
}
