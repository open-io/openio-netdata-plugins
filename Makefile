GOPATH ?= ${HOME}/go/

build:
		go mod download
		go build $(GOARGS) cmd/command.plugin/command.plugin.go
		go build $(GOARGS) cmd/container.plugin/container.plugin.go
		go build $(GOARGS) cmd/fs.plugin/fs.plugin.go
		go build $(GOARGS) cmd/openio.plugin/openio.plugin.go
		go build $(GOARGS) cmd/s3roundtrip.plugin/s3roundtrip.plugin.go
		go build $(GOARGS) cmd/zookeeper.plugin/zookeeper.plugin.go

test:
		go test -v $(GOARGS) ./...

check:
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin v1.28.3
		${GOPATH}/bin/golangci-lint run

ci: test check

format:
		if [ "$(shell gofmt -l . | wc -l)" -ne 0 ]; then \
					echo "golangci issues:"; \
							gofmt -l -d .; \
								fi
