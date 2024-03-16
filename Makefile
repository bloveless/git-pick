PATH = $(shell echo $$PATH)
BUILDPATH = $(shell realpath ./build/)

build: validate-path
	go build -o build/git-pick ./...

validate-path:
	@if [[ ":$(PATH):" == *":$(BUILDPATH):"* ]]; then \
		echo BUILDPATH was found in PATH; \
		exit 0; \
	else \
		echo BUILDPATH was not found in PATH; \
		echo Running \'export PATH=\""\$$PATH:\$$(pwd)/build\""\' should do what you need; \
		exit 1; \
	fi
