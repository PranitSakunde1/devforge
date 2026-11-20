package cmd

import (
	"devforge/internal/report"
	"fmt"
)

func generateReport(dir string) error {
	path, err := report.Generate(dir)
	if err != nil {
		return err
	}
	fmt.Println("HTML report generated:", path)
	return nil
}
