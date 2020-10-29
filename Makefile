IMAGE ?= storageos/init:test
GO_BUILD_CMD = go build -v
GO_ENV = GOOS=linux CGO_ENABLED=0

# Test node image name.
NODE_IMAGE ?= storageos/node:1.4.0
# Test scripts path.
SCRIPTS_PATH ?= /scripts

all: test build

.PHONY: build

build:
	@echo "Building init"
	$(GO_ENV) $(GO_BUILD_CMD) -o ./build/_output/bin/init .

tidy:
	go mod tidy -v
	go mod vendor -v

# Build the docker image
docker-build::
	docker build --no-cache . -f Dockerfile -t $(IMAGE)

# Push the docker image
docker-push:
	docker push ${IMAGE}

# Run tests
test: generate fmt vet
	go test -v -race `go list -v ./...`

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

clean:
	rm -rf build/_output

# Run the init scripts on the host.
run:
	docker run --rm \
		--cap-add=SYS_ADMIN \
		--privileged \
		-v /lib/modules:/lib/modules \
		-v /var/lib/storageos:/var/lib/storageos:rshared \
		-v /sys:/sys:rshared \
		storageos/init:test \
		/init -scripts=$(SCRIPTS_PATH) -nodeImage=$(NODE_IMAGE)

# Generate mocks.
generate: mockgen
	@PATH=$$(go env GOPATH)/bin:$(PATH); \
	go generate ./...

# Install mockgen dependency.
mockgen:
	@PATH=$$(go env GOPATH)/bin:$(PATH); \
	if ! which mockgen 1>/dev/null; then \
		go get github.com/golang/mock/gomock; \
		go get github.com/golang/mock/mockgen; \
	fi

# Prepare the repo for a new release. Run:
#   NEW_VERSION=<version> make release
release:
	sed -i -e "s/version=.*/version=\"$(NEW_VERSION)\" \\\/g" Dockerfile
