# Release and validation checklist

## Local Go checks
```powershell
go mod tidy
gofmt -w .
go test ./...
go vet ./...
go build -o devforge.exe .
```

## Compose validation
Run on a Windows host with Docker Desktop and Compose Watch support:
```powershell
devforge init
docker compose -f .\compose.yaml config
devforge doctor
devforge up
devforge watch
```
Edit one file in each service and confirm its endpoint remains healthy after reload.

## Test/report validation
```powershell
devforge test
devforge security
devforge report
Get-ChildItem .\reports -Recurse
```
Expected artifacts include JUnit XML, coverage files, Trivy JSON/SARIF and `reports/html/index.html`.

## Windows installation behavior
`devforge install` uses winget and does not self-elevate. Docker Desktop may require administrator approval and a restart. After installation, the user must launch Docker Desktop and select Linux containers. DevForge intentionally does not attempt to change system-wide virtualization or security settings.

## Release
Tag a commit with `vX.Y.Z`. GitHub Actions builds Windows amd64 and arm64 executables. Before publishing a release, add code signing and attach SHA256 checksums.
