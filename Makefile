#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh

test_coverage:
	docker-compose -f tests/env/docker-compose.yaml up -d --remove-orphans
	rm -rf coverage-ci
	mkdir ./coverage-ci
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/ws_origin.out -covermode=atomic ./websockets
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/http_config.out -covermode=atomic ./http/config
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/server_cmd.out -covermode=atomic ./server
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/struct_jobs.out -covermode=atomic ./jobs/job
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/pipeline_jobs.out -covermode=atomic ./jobs/pipeline
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/jobs_core.out -covermode=atomic ./tests/plugins/jobs
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/kv_plugin.out -covermode=atomic ./tests/plugins/kv
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/grpc_codec.out -covermode=atomic ./grpc/codec
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/grpc_parser.out -covermode=atomic ./grpc/parser
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/broadcast_plugin.out -covermode=atomic ./tests/plugins/broadcast
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/websockets.out -covermode=atomic ./tests/plugins/websockets
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/http.out -covermode=atomic ./tests/plugins/http
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/grpc_plugin.out -covermode=atomic ./tests/plugins/grpc
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/informer.out -covermode=atomic ./tests/plugins/informer
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/reload.out -covermode=atomic ./tests/plugins/reload
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/server.out -covermode=atomic ./tests/plugins/server
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/service.out -covermode=atomic ./tests/plugins/service
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/status.out -covermode=atomic ./tests/plugins/status
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/config.out -covermode=atomic ./tests/plugins/config
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/gzip.out -covermode=atomic ./tests/plugins/gzip
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/headers.out -covermode=atomic ./tests/plugins/headers
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/logger.out -covermode=atomic ./tests/plugins/logger
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/metrics.out -covermode=atomic ./tests/plugins/metrics
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/resetter.out -covermode=atomic ./tests/plugins/resetter
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/rpc.out -covermode=atomic ./tests/plugins/rpc
	echo 'mode: atomic' > ./coverage-ci/summary.txt
	tail -q -n +2 ./coverage-ci/*.out >> ./coverage-ci/summary.txt
	docker-compose -f tests/env/docker-compose.yaml down

test: ## Run application tests
	docker compose -f tests/env/docker-compose.yaml up -d --remove-orphans
	sleep 10
	go test -v -race -tags=debug ./jobs/pipeline
	go test -v -race -tags=debug ./http/config
	go test -v -race -tags=debug ./server
	go test -v -race -tags=debug ./jobs/job
	go test -v -race -tags=debug ./websockets
	go test -v -race -tags=debug ./grpc/codec
	go test -v -race -tags=debug ./grpc/parser
	go test -v -race -tags=debug ./tests/plugins/jobs
	go test -v -race -tags=debug ./tests/plugins/kv
	go test -v -race -tags=debug ./tests/plugins/broadcast
	go test -v -race -tags=debug ./tests/plugins/websockets
	go test -v -race -tags=debug ./tests/plugins/http
	go test -v -race -tags=debug ./tests/plugins/informer
	go test -v -race -tags=debug ./tests/plugins/reload
	go test -v -race -tags=debug ./tests/plugins/websockets
	go test -v -race -tags=debug ./tests/plugins/grpc
	go test -v -race -tags=debug ./tests/plugins/server
	go test -v -race -tags=debug ./tests/plugins/service
	go test -v -race -tags=debug ./tests/plugins/status
	go test -v -race -tags=debug ./tests/plugins/config
	go test -v -race -tags=debug ./tests/plugins/gzip
	go test -v -race -tags=debug ./tests/plugins/headers
	go test -v -race -tags=debug ./tests/plugins/logger
	go test -v -race -tags=debug ./tests/plugins/metrics
	go test -v -race -tags=debug ./tests/plugins/resetter
	go test -v -race -tags=debug ./tests/plugins/rpc
	docker compose -f tests/env/docker-compose.yaml down

generate-proto:
	protoc --proto_path=./internal/proto/jobs/v1beta --go_out=./internal/proto/jobs/v1beta jobs.proto
	protoc --proto_path=./internal/proto/kv/v1beta --go_out=./internal/proto/kv/v1beta kv.proto
	protoc --proto_path=./internal/proto/websockets/v1beta --go_out=./internal/proto/websockets/v1beta websockets.proto