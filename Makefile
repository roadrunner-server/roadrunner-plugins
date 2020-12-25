#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh

check: ## Run application linters
	cd checker && golangci-lint run -c ../.golangci.yml && cd ..
	cd config && golangci-lint run -c ../.golangci.yml && cd ..
	cd gzip && golangci-lint run -c ../.golangci.yml && cd ..
	cd headers && golangci-lint run -c ../.golangci.yml && cd ..
	cd logger && golangci-lint run -c ../.golangci.yml && cd ..
	cd metrics && golangci-lint run -c ../.golangci.yml && cd ..
	cd redis && golangci-lint run -c ../.golangci.yml && cd ..
	cd reload && golangci-lint run -c ../.golangci.yml && cd ..
	cd resetter && golangci-lint run -c ../.golangci.yml && cd ..
	cd rpc && golangci-lint run -c ../.golangci.yml && cd ..
	cd static && golangci-lint run -c ../.golangci.yml && cd ..
