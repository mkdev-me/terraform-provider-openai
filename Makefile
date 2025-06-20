TEST?=$$(go list ./... | grep -v 'vendor')
SWEEP?=all
SWEEP_DIR?=./internal
HOSTNAME=github.com
NAMESPACE=fjcorp
NAME=openai
BINARY=terraform-provider-${NAME}
VERSION=0.1.0
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
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

# For Terraform 0.12.x compatibility
install012: build
	mkdir -p ~/.terraform.d/plugins/${OS_ARCH}/
	cp ${BINARY} ~/.terraform.d/plugins/${OS_ARCH}/terraform-provider-${NAME}_v${VERSION}

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
	rm -vrf examples/.terraform /tmp/terraform-provider-openai.log

# These targets are used for running/testing the provider with example configurations
apply: install _terraform_cleanup _terraform_init _terraform_apply _terraform_log
plan: install _terraform_cleanup _terraform_init _terraform_plan _terraform_log
destroy: install _terraform_cleanup _terraform_init _terraform_destroy _terraform_log

_terraform_cleanup:
	# Cleanup last run
	rm -vrf examples/.terraform /tmp/terraform-provider-openai.log
_terraform_init:
	# Initialize terraform
	(cd examples && terraform init)
_terraform_log:
	# Print the debug log
	cat /tmp/terraform-provider-openai.log 2>/dev/null || echo "No log file found"
_terraform_apply:
	(cd examples && OPENAI_PROVIDER_DEBUG=1 terraform apply) || true
_terraform_plan:
	(cd examples && OPENAI_PROVIDER_DEBUG=1 terraform plan)
_terraform_destroy:
	(cd examples && OPENAI_PROVIDER_DEBUG=1 terraform destroy)

# Test with different Terraform versions using Docker
terraform012:
	docker build . --build-arg version=0.12.29 -t terraform-0.12-provider-openai
	docker run terraform-0.12-provider-openai plan

terraform013:
	docker build . --build-arg version=0.13.5 -t terraform-0.13-provider-openai
	docker run terraform-0.13-provider-openai plan

terraform014:
	docker build . --build-arg version=0.14.2 -t terraform-0.14-provider-openai
	docker run terraform-0.14-provider-openai plan

.PHONY: build fmt lint test testacc install install012 clean release plan apply destroy _terraform_cleanup _terraform_init _terraform_log _terraform_apply _terraform_plan _terraform_destroy 