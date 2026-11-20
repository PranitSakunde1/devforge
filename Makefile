.PHONY: tidy fmt test vet build windows clean

tidy:
	go mod tidy

fmt:
	gofmt -w .

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -o devforge .

windows:
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/devforge-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/devforge-windows-arm64.exe .

clean:
	rm -rf dist
