## For usage, see README.md

PKGS := ./cli ./cli/distrilock ./api ./api/client ./api/core ./api/client/tcp ./cli/distrilock-ws ./api/client/ws
PKG := bitbucket.org/gdm85/go-distrilock

all: vendor build test

setup: vendor godoc-tool codeqa-tools
	@echo "Setup completed"

vendor:
	if ! ls vendor/github.com/ogier/pflag/* 2>/dev/null >/dev/null; then git submodule update --init --recursive; fi

build:
	mkdir -p bin
	GOBIN="$(CURDIR)/bin" go install $(PKGS)

run: build
	bin/distrilock

test:
	scripts/run-tests.sh $(PKGS)

benchmark:
	scripts/run-tests.sh -bench=. -benchtime=1s $(PKGS)

race:
	scripts/run-tests.sh -race $(PKGS)

simplify:
	gofmt -w -s $(shell find $(PKGS) -name '*.go' -type f)

godoc: godoc-tool
	@echo "Go documentation available at: http://localhost:8080/pkg/$(PKG)/"
	godoc -http=:8080

godoc-static:
	rm -rf docs
	mkdir -p docs
	scripts/gen-godoc.sh $(PKG) docs

godoc-tool:
	go get golang.org/x/tools/cmd/godoc

codeqa-tools:
	go get github.com/golang/lint/golint github.com/kisielk/errcheck

codeqa: vet lint errcheck

vet:
	@echo -n "**** Running: "
	go vet $(PKGS) || true

lint:
	@echo -n "**** Running: "
	golint $(PKGS)

errcheck:
	@echo -n "**** Running: "
	errcheck -ignorepkg os $(PKGS)

clean:
	rm -rf bin/ docs/

.PHONY: all build test clean godoc errcheck codeqa codeqa-tools vet lint godoc-tool godoc-static vendor benchmark race setup
