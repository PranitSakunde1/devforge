package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateHTMLWithJUnitAndTrivy(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "reports", "junit"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "reports", "security"), 0755); err != nil {
		t.Fatal(err)
	}
	junit := `<testsuite name="unit" tests="3" failures="1" errors="0" skipped="0"></testsuite>`
	trivy := `{"Results":[{"Target":"image","Vulnerabilities":[{"Severity":"High"},{"Severity":"LOW"}]}]}`
	_ = os.WriteFile(filepath.Join(dir, "reports", "junit", "unit.xml"), []byte(junit), 0644)
	_ = os.WriteFile(filepath.Join(dir, "reports", "security", "trivy-fs.json"), []byte(trivy), 0644)
	out, err := Generate(dir)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	html := string(b)
	for _, want := range []string{"DevForge test &amp; security report", "unit", "High", "3", "1"} {
		if !strings.Contains(html, want) {
			t.Fatalf("report missing %q", want)
		}
	}
}
