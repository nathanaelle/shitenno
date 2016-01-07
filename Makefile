#!/usr/bin/make -f
# -*- makefile -*-

# basic Makefile
OS=linux
ARCH=amd64

export	GOOS=$(shell [ "x${OS}" != "x" ] && echo ${OS} || (go env GOOS) )
export	GOARCH=$(shell [ "x${ARCH}" != "x" ] && echo ${ARCH} || (go env GOARCH) )
export	GO15VENDOREXPERIMENT=1

.PHONY: build

all: update build

update:

build:
	@echo building for ${GOOS}/${GOARCH}${GOARM}
	go build -o shitenno src/*.go
