package report

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type junitNode struct {
	Tests    int         `xml:"tests,attr"`
	Failures int         `xml:"failures,attr"`
	Errors   int         `xml:"errors,attr"`
	Skipped  int         `xml:"skipped,attr"`
	Name     string      `xml:"name,attr"`
	Suites   []junitNode `xml:"testsuite"`
}
type junitSuite struct {
	Tests, Failures, Errors, Skipped int
	Name                             string
}
type coverageItem struct {
	Name    string
	Percent float64
}
type trivyReport struct {
	Results []struct {
		Vulnerabilities []struct {
			Severity string `json:"Severity"`
		} `json:"Vulnerabilities"`
		Misconfigurations []struct {
			Severity string `json:"Severity"`
		} `json:"Misconfigurations"`
		Secrets []struct {
			Severity string `json:"Severity"`
		} `json:"Secrets"`
	} `json:"Results"`
}
type Summary struct {
	Suites                           []junitSuite
	Coverage                         []coverageItem
	Tests, Failures, Errors, Skipped int
	Critical, High, Medium, Low      int
	TrivyAvailable                   bool
	Generated                        string
}

var page = template.Must(template.New("report").Parse(`<!doctype html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>DevForge report</title><style>
body{font:15px system-ui;margin:36px auto;max-width:1000px;padding:0 20px;color:#20242b}
h1{margin-bottom:4px}.muted{color:#687080}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(130px,1fr));gap:12px;margin:24px 0}
.card{border:1px solid #dfe3e8;border-radius:12px;padding:16px}.value{font-size:28px;font-weight:700}
table{border-collapse:collapse;width:100%}td,th{text-align:left;padding:12px;border-bottom:1px solid #e5e7eb}
.good{color:#087443}.bad{color:#b42318}
</style></head><body><h1>DevForge test &amp; security report</h1><p class="muted">Generated {{.Generated}}</p>
<div class="grid">
<div class="card"><div class="muted">Tests</div><div class="value">{{.Tests}}</div></div>
<div class="card"><div class="muted">Failures</div><div class="value {{if .Failures}}bad{{else}}good{{end}}">{{.Failures}}</div></div>
<div class="card"><div class="muted">Errors</div><div class="value">{{.Errors}}</div></div>
<div class="card"><div class="muted">Skipped</div><div class="value">{{.Skipped}}</div></div>
</div><h2>Test suites</h2><table><tr><th>Suite</th><th>Tests</th><th>Failures</th><th>Errors</th><th>Skipped</th></tr>
{{range .Suites}}<tr><td>{{.Name}}</td><td>{{.Tests}}</td><td>{{.Failures}}</td><td>{{.Errors}}</td><td>{{.Skipped}}</td></tr>{{else}}<tr><td colspan="5">No JUnit XML files found. Run devforge test first.</td></tr>{{end}}</table>
<h2>Coverage</h2><table><tr><th>Source</th><th>Coverage</th></tr>
{{range .Coverage}}<tr><td>{{.Name}}</td><td>{{printf "%.1f" .Percent}}%</td></tr>{{else}}<tr><td colspan="2">No coverage artifacts found. Run devforge test first.</td></tr>{{end}}</table>
<h2>Trivy findings</h2>{{if .TrivyAvailable}}<div class="grid">
<div class="card"><div class="muted">Critical</div><div class="value">{{.Critical}}</div></div>
<div class="card"><div class="muted">High</div><div class="value">{{.High}}</div></div>
<div class="card"><div class="muted">Medium</div><div class="value">{{.Medium}}</div></div>
<div class="card"><div class="muted">Low</div><div class="value">{{.Low}}</div></div>
</div>{{else}}<p class="muted">No Trivy JSON report found. Run devforge security first.</p>{{end}}
<p class="muted">Raw XML, JSON, coverage and SARIF artifacts are retained under reports/.</p></body></html>`))

func Generate(workspace string) (string, error) {
	reportDir := filepath.Join(workspace, "reports")
	s := Summary{Generated: time.Now().Format(time.RFC3339)}
	_ = filepath.Walk(reportDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		lower := strings.ToLower(path)
		if strings.HasSuffix(lower, ".xml") && strings.Contains(lower, string(filepath.Separator)+"junit"+string(filepath.Separator)) {
			b, e := os.ReadFile(path)
			if e == nil {
				var node junitNode
				if xml.Unmarshal(b, &node) == nil {
					if node.Name != "" {
						addSuite(&s, node)
					} else {
						for _, child := range node.Suites {
							addNode(&s, child)
						}
					}
				}
			}
		}
		if strings.HasSuffix(lower, ".xml") && strings.Contains(lower, string(filepath.Separator)+"coverage"+string(filepath.Separator)) {
			if item, ok := readCoverageXML(path); ok {
				s.Coverage = append(s.Coverage, item)
			}
		}
		if strings.HasSuffix(lower, ".out") && strings.Contains(lower, string(filepath.Separator)+"coverage"+string(filepath.Separator)) {
			if item, ok := readGoCoverage(path); ok {
				s.Coverage = append(s.Coverage, item)
			}
		}
		return nil
	})
	for _, name := range []string{"trivy-fs.json", "trivy-image.json"} {
		b, e := os.ReadFile(filepath.Join(reportDir, "security", name))
		if e != nil {
			continue
		}
		var tr trivyReport
		if json.Unmarshal(b, &tr) != nil {
			continue
		}
		s.TrivyAvailable = true
		for _, r := range tr.Results {
			for _, v := range r.Vulnerabilities {
				addSeverity(&s, v.Severity)
			}
			for _, v := range r.Misconfigurations {
				addSeverity(&s, v.Severity)
			}
			for _, v := range r.Secrets {
				addSeverity(&s, v.Severity)
			}
		}
	}
	out := filepath.Join(reportDir, "html", "index.html")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return "", err
	}
	f, err := os.Create(out)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := page.Execute(f, s); err != nil {
		return "", err
	}
	return out, nil
}
func addNode(s *Summary, n junitNode) {
	if n.Name != "" {
		addSuite(s, n)
	}
	for _, child := range n.Suites {
		addNode(s, child)
	}
}
func addSuite(s *Summary, n junitNode) {
	name := n.Name
	if name == "" {
		name = "unnamed suite"
	}
	s.Suites = append(s.Suites, junitSuite{n.Tests, n.Failures, n.Errors, n.Skipped, name})
	s.Tests += n.Tests
	s.Failures += n.Failures
	s.Errors += n.Errors
	s.Skipped += n.Skipped
}
func readCoverageXML(path string) (coverageItem, bool) {
	b, e := os.ReadFile(path)
	if e != nil {
		return coverageItem{}, false
	}
	var doc struct {
		LineRate string `xml:"line-rate,attr"`
		Counters []struct {
			Type    string `xml:"type,attr"`
			Missed  int    `xml:"missed,attr"`
			Covered int    `xml:"covered,attr"`
		} `xml:"counter"`
	}
	if xml.Unmarshal(b, &doc) != nil {
		return coverageItem{}, false
	}
	if doc.LineRate != "" {
		v, e := strconv.ParseFloat(doc.LineRate, 64)
		if e == nil {
			return coverageItem{filepath.Base(path), v * 100}, true
		}
	}
	for _, c := range doc.Counters {
		if c.Type == "LINE" {
			total := c.Missed + c.Covered
			if total > 0 {
				return coverageItem{filepath.Base(filepath.Dir(path)), float64(c.Covered) * 100 / float64(total)}, true
			}
		}
	}
	return coverageItem{}, false
}
func readGoCoverage(path string) (coverageItem, bool) {
	b, e := os.ReadFile(path)
	if e != nil {
		return coverageItem{}, false
	}
	var total, covered float64
	for _, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 3 || fields[0] == "mode:" {
			continue
		}
		count, e := strconv.Atoi(fields[2])
		if e != nil {
			continue
		}
		stmts, e := strconv.Atoi(fields[1])
		if e != nil {
			continue
		}
		total += float64(stmts)
		if count > 0 {
			covered += float64(stmts)
		}
	}
	if total == 0 {
		return coverageItem{}, false
	}
	return coverageItem{"Go", covered * 100 / total}, true
}
func addSeverity(s *Summary, level string) {
	switch strings.ToUpper(level) {
	case "CRITICAL":
		s.Critical++
	case "HIGH":
		s.High++
	case "MEDIUM":
		s.Medium++
	case "LOW":
		s.Low++
	}
}
func Explain(path string) { fmt.Println("HTML report:", path) }
