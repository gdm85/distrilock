## For usage, see README.md

PKGS := ./cli/distrilock ./api ./api/dlclient
PKG := bitbucket.org/gdm85/go-distrilock

all: vendor build test

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
	gofmt -w -s cli/distrilock/*.go api/*.go api/dlclient/*.go

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
	go vet $(PKGS)

lint:
	golint $(PKGS)

errcheck:
	errcheck -ignorepkg os $(PKGS)

clean:
	rm -rf bin/ docs/

.PHONY: all build test clean godoc errcheck codeqa codeqa-tools vet lint godoc-tool godoc-static vendor benchmark race
