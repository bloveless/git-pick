.PHONY: build
build:
	go build -o build/git-pick ./...

.PHONY: build-amd64
build-amd64:
	GOOS=darwin GOARCH=amd64 go build -o build/git-pick-amd64 ./...

install: build
	cp build/git-pick ~/.local/bin
