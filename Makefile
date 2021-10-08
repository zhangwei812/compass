# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: atlas android ios atlas-cross evm all test clean
.PHONY: atlas-linux atlas-linux-386 atlas-linux-amd64 atlas-linux-mips64 atlas-linux-mips64le
.PHONY: atlas-linux-arm atlas-linux-arm-5 atlas-linux-arm-6 atlas-linux-arm-7 atlas-linux-arm64
.PHONY: atlas-darwin atlas-darwin-386 atlas-darwin-amd64
.PHONY: atlas-windows atlas-windows-386 atlas-windows-amd64

GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run

atlas:
	$(GORUN) build/ci.go install .
	@echo "Done building."
	@echo "Run \"$(GOBIN)/atlas\" to launch atlas."

