all: build

run: assets
	go run *.go

install: assets
	go install

deps:
	go get -u golang.org/x/sys/...
	go get

assets:
	go-bindata defaultTemplate

build: assets
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -a -tags netgo -ldflags '-w' -o bin/blogme-linux-arm7 *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -tags netgo -ldflags '-w' -o bin/blogme-linux-arm64 *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o bin/blogme-linux-amd64 *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -a -tags netgo -ldflags '-w' -o bin/blogme-linux-x86 *.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o bin/blogme-windows-amd64.exe *.go
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -a -tags netgo -ldflags '-w' -o bin/blogme-windows-x86.exe *.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o bin/blogme-darwin-amd64 *.go

clear: 
	rm bin/* -Rf
