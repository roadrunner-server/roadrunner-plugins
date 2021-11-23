PHP_OUT_DIR = ../../../../src
GO_OUT_DIR = ./

all: build
build: health_go health_php

health_php:
	protoc --php_out=$(PHP_OUT_DIR) --php-grpc_out=$(PHP_OUT_DIR) \
    	health.proto
health_go:
	protoc --go_out=$(GO_OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT_DIR) --go-grpc_opt=paths=source_relative \
		health.proto
clean:
	rm -rf ./$(PHP_OUT_DIR)/*
	rm ./$(GO_OUT_DIR)/*.go
