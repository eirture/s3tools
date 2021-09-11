

VERSION ?= unknown
BUILD_DATE = $(shell date +%Y-%m-%d)

GO_LDFLAGS := -X github.com/eirture/s3tools/pkg/build.Version=$(VERSION) $(GO_LDFLAGS)
GO_LDFLAGS := -X github.com/eirture/s3tools/pkg/build.Date=$(BUILD_DATE) $(GO_LDFLAGS)

.PHONY: build
build:
	@mkdir -p bin
	go build -ldflags "${GO_LDFLAGS}" -o ./bin/ ./pkg/cmd/s3tools

.PHONY: clean
clean:
	rm -rf bin/*
