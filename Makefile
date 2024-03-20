PATH = $(shell echo $$PATH)
BUILDPATH = $(shell realpath ./build/)

build: validate-path
	go build -o build/git-pick ./...

build-amd64: validate-path
	GOOS=darwin GOARCH=amd64 go build -o build/git-pick-amd64 ./...

validate-path:
	@if [[ ":$(PATH):" == *":$(BUILDPATH):"* ]]; then \
		echo BUILDPATH was found in PATH; \
		exit 0; \
	else \
		exit 1; \
	fi
