install-dep:
	@go get -d -u github.com/rakyll/statik
	@go get -u -d ./...

build:
	@echo "Building the LWN Simulator"
	@echo "Building the User Interface"
	@cd webserver && statik -src=public -f 1
	@if not exist bin mkdir bin
	@xcopy "cmd/c2.json" "bin" /F
	@xcopy "cmd/datasamples" "bin/datasamples" /E /I
	@echo "Building the source"
	@go build -o bin/lwnsimulator cmd/main.go
	@mingw32-make build-x64
	@mingw32-make build-x86
	@mingw32-make build-windows
	@echo "Build Complete"

build-x64:
	@set GOOS=linux
	@set GOARCH=amd64
	@go build -o bin/lwnsimulators_x64 cmd/main.go

build-x86:
	@set GOOS=linux
	@set GOARCH=386
	@go build -o bin/lwnsimulators_x86 cmd/main.go

build-windows:
	@set GOOS=windows
	@set GOARCH=amd64
	@go build -o bin/lwnsimulator.exe cmd/main.go

run:
	@go run cmd/main.go

run-release:
	@cd bin
	@lwnsimulator.exe
