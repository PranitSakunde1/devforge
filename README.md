# DevForge

Windows-first Docker development environment CLI for Go, Java 21, Python, Redis, hot reload, test orchestration and Trivy.

## Requirements
- Windows 10/11 (Windows 11 recommended)
- Docker Desktop configured for Linux containers
- Docker Compose v2.23+ (required for the configured Watch actions)
- Go 1.23+ to build DevForge

The CLI runs natively on Windows. Development toolchains and infrastructure run in Linux containers. WSL is not a user prerequisite. `devforge up` builds Go/Java/Python images and pulls Redis; `devforge security` pulls Trivy on first use.

## Build
```powershell
go mod tidy
go build -o devforge.exe .
```

## First run
```powershell
.\devforge.exe install
# Start Docker Desktop, wait for Linux engine, then:
.\devforge.exe doctor
.\devforge.exe init
.\devforge.exe up
.\devforge.exe watch
```

## Daily commands
```powershell
devforge logs
devforge test
devforge security
devforge security --image my-service:dev
devforge report
devforge down
devforge clean                 # preserves named volumes
devforge clean --volumes       # destructive; requires typing DELETE
```

## Ports
- Go: http://localhost:8081/health
- Java: http://localhost:8082/health
- Python: http://localhost:8083/health
- Redis: localhost:6379

## Notes
- `install` uses winget for Git and Docker Desktop. It does not install runtimes on Windows.
- Docker Desktop installation may require an administrator-approved installer and a reboot. DevForge does not self-elevate.
- Compose Watch requires a sufficiently recent Docker Compose v2 release.
- Generated files are preserved by `generate`; use `generate --force` only after reviewing changes.
- Test outputs and Trivy JSON/SARIF are retained under `reports/`.

## Build and release

```powershell
go mod tidy
go test ./...
go vet ./...
go build -o devforge.exe .
```

Push a semantic version tag (for example `v0.2.0`) to trigger CI tests and Windows amd64/arm64 release builds. The workflow attaches SHA256 checksums. Code signing is not configured yet; unsigned executables may trigger SmartScreen warnings.
