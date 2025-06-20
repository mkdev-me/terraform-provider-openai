TEST?=$$(go list ./... | grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
# For local development
DEV_HOSTNAME=github.com
# For Terraform Registry
REGISTRY_HOSTNAME=registry.terraform.io
NAMESPACE=fjcorp
NAME=openai
BINARY=terraform-provider-${NAME}
VERSION=1.0.0
OS_ARCH=darwin_arm64

default: install

build:
	go build -o ${BINARY}

release:
	mkdir -p ./bin
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=darwin GOARCH=arm64 go build -o ./bin/${BINARY}_${VERSION}_darwin_arm64
	GOOS=freebsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_windows_amd64

install: build
	mkdir -p ~/.terraform.d/plugins/${REGISTRY_HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY} ~/.terraform.d/plugins/${REGISTRY_HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

install-dev: build
	mkdir -p ~/.terraform.d/plugins/${DEV_HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY} ~/.terraform.d/plugins/${DEV_HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w $(GOFMT_FILES)

lint:
	@echo "==> Checking source code against linters..."
	golangci-lint run ./...

test:
	go test -covermode=atomic -coverprofile=coverage.out $(TEST) || exit 1

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

clean:
	rm -f ${BINARY}
	rm -rf ./bin

.PHONY: build fmt lint test testacc install install-dev clean release 