IMAGE ?= storageos/init:test
GO_BUILD_CMD = go build -v
GO_ENV = GOOS=linux CGO_ENABLED=0

MPIMAGE	?= docker.io/soegarots/mpinit:latest

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
docker-build:
	docker build --no-cache . -f Dockerfile -t $(IMAGE)

# Push the docker image
docker-push:
	docker push ${IMAGE}

SUDO		= sudo
BUILDAH		= $(SUDO) buildah
PODMAN		= $(SUDO) podman

BUILD_DEPS			=

BUILD_ARGS_COMMON	=
BUILD_ARGS_ARM64	= $(BUILD_ARGS_COMMON)
BUILD_ARGS_AMD64	= $(BUILD_ARGS_COMMON)

BUILDAH_MANIFEST	= mpbuild-init
BUILDAH_URI			= docker://$(MPIMAGE)

mpbuild: mpbuild-prep mpbuild-amd64 mpbuild-arm64

mprelease: mpbuild-finalise

mpbuild-prep:
	$(BUILDAH) manifest rm $(BUILDAH_MANIFEST) || true
	$(BUILDAH) manifest create $(BUILDAH_MANIFEST)

mpbuild-amd64: $(BUILD_DEPS)
	$(BUILDAH) build --tag $(MPIMAGE) $(BUILD_ARGS_AMD64) --manifest $(BUILDAH_MANIFEST) --arch amd64 .

mpbuild-arm64: $(BUILD_DEPS)
	$(BUILDAH) build --tag $(MPIMAGE) $(BUILD_ARGS_ARM64) --manifest $(BUILDAH_MANIFEST) --arch arm64 .

mpbuild-finalise:
	$(BUILDAH) manifest push --all $(BUILDAH_MANIFEST) $(BUILDAH_URI)


cache-build-amd64: $(BUILD_DEPS)
	$(PODMAN) build --arch amd64 $(BUILD_ARGS_AMD64) -t deleteme:latest .

cache-build-arm64: $(BUILD_DEPS)
	$(PODMAN) build --arch arm64 $(BUILD_ARGS_ARM64) -t deleteme:latest .


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
