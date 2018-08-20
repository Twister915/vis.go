.PHONY: build

build: clean
	@go build -o viz ./cmd/fft_gl

format:
	@go fmt $(shell go list ./...)
	@goimports -w $(shell find . -name "*.go" -type f | grep -v vendor | xargs)

clean:
	@rm -f viz

cloc:
	cloc --exclude_dir vendor,.idea,.git .