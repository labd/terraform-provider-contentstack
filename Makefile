.PHONY: docs
LOCAL_TEST_VERSION = 99.0.0
OS_ARCH = darwin_arm64

build:
	go build


# Build local provider with very high version number for easier local testing and debugging
# see: https://discuss.hashicorp.com/t/easiest-way-to-use-a-local-custom-provider-with-terraform-0-13/12691/5
build-local:
	go build -o terraform-provider-contentstack_${LOCAL_TEST_VERSION}
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/labd/contentstack/${LOCAL_TEST_VERSION}/${OS_ARCH}
	cp terraform-provider-contentstack_${LOCAL_TEST_VERSION} ~/.terraform.d/plugins/registry.terraform.io/labd/contentstack/${LOCAL_TEST_VERSION}/${OS_ARCH}/terraform-provider-contentstack_v${LOCAL_TEST_VERSION}

lint:
	staticcheck ./...

format:
	go fmt ./...

test:
	go test -v ./...

docs:
	tfplugindocs

coverage-html:
	go test -race -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... ./...
	go tool cover -html=coverage.txt

coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... ./...
	go tool cover -func=coverage.txt
