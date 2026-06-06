HOSTNAME=registry.opentofu.org
NAMESPACE=rinseaid
NAME=technitium-dns
BINARY=terraform-provider-${NAME}
VERSION=0.1.0
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

default: build

build:
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test ./... -v -count=1

testcover:
	go test ./... -v -count=1 -coverprofile=coverage-unit.out -covermode=atomic
	go tool cover -func=coverage-unit.out

testacc:
	TF_ACC=1 go test ./... -v -count=1 -timeout 30m

testacccover:
	TF_ACC=1 go test ./... -v -count=1 -timeout 30m -coverprofile=coverage-acc.out -covermode=atomic
	go tool cover -func=coverage-acc.out

fmt:
	go fmt ./...
	terraform fmt -recursive examples/

lint:
	golangci-lint run ./...

docs:
	go generate ./...
	@echo "Docs generated in docs/"

clean:
	rm -f ${BINARY}

.PHONY: default build install test testcover testacc testacccover fmt lint docs clean
