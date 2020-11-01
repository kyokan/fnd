fpm = @docker run --rm -i -v "$(CURDIR):$(CURDIR)" -w "$(CURDIR)" -u $(shell id -u) digitalocean/fpm:latest
git_commit := $(shell git log -1 --format='%H')
git_tag := $(shell git describe --tags --abbrev=0)
ldflags = -X fnd/version.GitCommit=$(git_commit) -X fnd/version.GitTag=$(git_tag)
build_flags := -ldflags '$(ldflags)'

all: fnd
.PHONY: all

all-cross:
	GOOS=darwin GOARCH=amd64 go build $(build_flags) -o ./build/fnd-darwin-amd64 ./cmd/fnd/main.go
	GOOS=darwin GOARCH=amd64 go build $(build_flags) -o ./build/fnd-cli-darwin-amd64 ./cmd/fnd-cli/main.go
	GOOS=linux GOARCH=amd64 go build $(build_flags) -o ./build/fnd-linux-amd64 ./cmd/fnd/main.go
	GOOS=windows GOARCH=amd64 go build $(build_flags) -o ./build/fnd.exe ./cmd/fnd/main.go
	GOOS=linux GOARCH=amd64 go build $(build_flags) -o ./build/fnd-cli-linux-amd64 ./cmd/fnd-cli/main.go
.PHONY: all-cross

fnd-cli: proto
	go build $(build_flags) -o ./build/fnd-cli ./cmd/fnd-cli/main.go
.PHONY: fnd-cli

fnd: proto
	go build $(build_flags) -o ./build/fnd ./cmd/fnd/main.go
.PHONY: fnd

proto:
	protoc -I rpc/v1/ rpc/v1/api.proto --go_out=plugins=grpc:rpc/v1
.PHONY: proto

test: proto
	go test ./... -v
.PHONY: test

install: all
	sudo mv ./build/fnd /usr/local/bin
	sudo mv ./build/fnd-cli /usr/local/bin
.PHONY: install

clean:
	rm -rf ./build
.PHONY: clean

fmt:
	go mod tidy
	goimports -w .
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs gofmt -w -s
.PHONY: fmt

package-deb: all-cross
	@$(fpm) --output-type deb \
		 --verbose \
		 --input-type dir \
		 --force \
		 --architecture amd64 \
		 --package ./build/fnd-$(version)-amd64.deb \
		 --no-depends \
		 --name fnd \
		 --maintainer "The Footnote Maintainers" \
		 --version $(version) \
		 --description "Footnote Network Daemon" \
		 --license MIT \
		 --vendor "The Footnote Maintainers" \
		 --url "https://fnd.network" \
		 --log info \
		 --deb-user fnd \
		 --deb-group nobody \
		 --after-install packaging/scripts/after_install.sh \
		 --after-remove packaging/scripts/after_remove.sh \
		 --config-files /lib/systemd/system/fnd.service \
		 build/fnd-linux-amd64=/usr/bin/fnd \
		 build/fnd-cli-linux-amd64=/usr/bin/fnd-cli \
		 packaging/lib/systemd/system/fnd.service=/lib/systemd/system/fnd.service
	@docker run --rm -i -v "$(CURDIR):$(CURDIR)" -w "$(CURDIR)" ubuntu:xenial /bin/bash -c 'dpkg --info ./build/fnd-$(version)-amd64.deb && dpkg -c ./build/fnd-$(version)-amd64.deb'
.PHONY: package-deb
