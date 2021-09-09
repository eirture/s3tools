

.PHONY: build
build:
	mkdir -p bin
	go build -o ./bin/ ./pkg/cmd/s3tools

clean:
	rm -rf bin/*
