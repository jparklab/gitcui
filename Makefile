
REV=v0.1.0

all: gitcui

gitcui:
	mkdir -p bin
	GO111MODULE=on GOOS=linux go build -a -ldflags '-X main.version=$(REV) -extldflags "-static"' -o ./bin/gitcui ./cmd/gitcui

test:
	GO111MODULE=on go test ./...
