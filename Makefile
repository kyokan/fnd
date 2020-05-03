fpm = @docker run --rm -i -v "$(CURDIR):$(CURDIR)" -w "$(CURDIR)" -u $(shell id -u) digitalocean/fpm:latest
git_commit := $(shell git log -1 --format='%H')
git_tag := $(shell git describe --tags --abbrev=0)
ldflags = -X ddrp/version.GitCommit=$(git_commit) -X ddrp/version.GitTag=$(git_tag)
build_flags := -ldflags '$(ldflags)'

all: ddrpd ddrpcli
.PHONY: all

all-cross:
	GOOS=darwin GOARCH=amd64 go build $(build_flags) -o ./build/ddrpd-darwin-amd64 ./cmd/ddrpd/main.go
	GOOS=darwin GOARCH=amd64 go build $(build_flags) -o ./build/ddrpcli-darwin-amd64 ./cmd/ddrpcli/main.go
	GOOS=linux GOARCH=amd64 go build $(build_flags) -o ./build/ddrpd-linux-amd64 ./cmd/ddrpd/main.go
	GOOS=linux GOARCH=amd64 go build $(build_flags) -o ./build/ddrpcli-linux-amd64 ./cmd/ddrpcli/main.go
.PHONY: all-cross

ddrpcli: proto
	go build $(build_flags) -o ./build/ddrpcli ./cmd/ddrpcli/main.go
.PHONY: ddrpcli

ddrpd: proto
	go build $(build_flags) -o ./build/ddrpd ./cmd/ddrpd/main.go
.PHONY: ddrpd

proto:
	protoc -I rpc/v1/ rpc/v1/api.proto --go_out=plugins=grpc:rpc/v1
.PHONY: proto

test: proto
	go test ./... -v
.PHONY: test

install: all
	sudo mv ./build/ddrpd /usr/local/bin
	sudo mv ./build/ddrpcli /usr/local/bin
.PHONY: install

clean:
	rm -rf ./build
.PHONY: clean

fmt:
	go mod tidy
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs gofmt -w -s
.PHONY: fmt

package-deb: all-cross
	@$(fpm) --output-type deb \
		 --verbose \
		 --input-type dir \
		 --force \
		 --architecture amd64 \
		 --package ./build/ddrp-$(version)-amd64.deb \
		 --no-depends \
		 --name ddrp \
		 --maintainer "The DDRP Maintainers" \
		 --version $(version) \
		 --description "DDRP Network Daemon" \
		 --license MIT \
		 --vendor "The DDRP Maintainers" \
		 --url "https://ddrp.network" \
		 --log info \
		 --deb-user ddrp \
		 --deb-group nobody \
		 --after-install packaging/scripts/after_install.sh \
		 --after-remove packaging/scripts/after_remove.sh \
		 --config-files /lib/systemd/system/ddrpd.service \
		 build/ddrpd-linux-amd64=/usr/bin/ddrpd \
		 build/ddrpcli-linux-amd64=/usr/bin/ddrpcli \
		 packaging/lib/systemd/system/ddrpd.service=/lib/systemd/system/ddrpd.service
	@docker run --rm -i -v "$(CURDIR):$(CURDIR)" -w "$(CURDIR)" ubuntu:xenial /bin/bash -c 'dpkg --info ./build/ddrp-$(version)-amd64.deb && dpkg -c ./build/ddrp-$(version)-amd64.deb'
.PHONY: package-deb