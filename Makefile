#!/usr/bin/make -f
# -*- makefile -*-

# basic Makefile
OS=linux
ARCH=amd64

export	GOOS=$(shell [ "x${OS}" != "x" ] && echo ${OS} || (go env GOOS) )
export	GOARCH=$(shell [ "x${ARCH}" != "x" ] && echo ${ARCH} || (go env GOARCH) )

.PHONY: build

all: update build

update:

build:
	@echo building for ${GOOS}/${GOARCH}${GOARM}
	go build -o shitenno.${GOOS} src/*.go
