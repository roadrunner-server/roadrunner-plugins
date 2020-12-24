#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh

test: ## Run application tests
	go test -v -race -cover -tags=debug -covermode=atomic ./checker
	go test -v -race -cover -tags=debug -covermode=atomic ./config
	go test -v -race -cover -tags=debug -covermode=atomic ./gzip
	go test -v -race -cover -tags=debug -covermode=atomic ./headers
	go test -v -race -cover -tags=debug -covermode=atomic ./http
	go test -v -race -cover -tags=debug -covermode=atomic ./informer
	go test -v -race -cover -tags=debug -covermode=atomic ./logger
	go test -v -race -cover -tags=debug -covermode=atomic ./metrics
	go test -v -race -cover -tags=debug -covermode=atomic ./redis
	go test -v -race -cover -tags=debug -covermode=atomic ./reload
	go test -v -race -cover -tags=debug -covermode=atomic ./resetter
	go test -v -race -cover -tags=debug -covermode=atomic ./rpc
	go test -v -race -cover -tags=debug -covermode=atomic ./server
	go test -v -race -cover -tags=debug -covermode=atomic ./static

test_http:
	go test -v -race -cover -covermode=atomic ./http