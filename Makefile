BINARY_NAME=oam-ecs
PACKAGES=./internal...
GOBIN=${PWD}/bin/tools
COVERAGE=coverage.out

DESTINATION=./bin/local/${BINARY_NAME}
VERSION=$(shell git describe --always --tags)

LINKER_FLAGS=-X github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/version.Version=${VERSION}
# RELEASE_BUILD_LINKER_FLAGS disables DWARF and symbol table generation to reduce binary size
RELEASE_BUILD_LINKER_FLAGS=-s -w

all: build

.PHONY: build
build: format packr-build compile-local packr-clean

.PHONY: stage-release
stage-release: packr-build compile-darwin compile-linux compile-windows packr-clean

.PHONY: format
format:
	go fmt ./...

packr-build: tools
	@echo "Packaging static files" &&\
	env -i PATH=$$PATH:${GOBIN} GOCACHE=$$(go env GOCACHE) GOPATH=$$(go env GOPATH) \
	go generate ./...

packr-clean: tools
	@echo "Cleaning up static files generated code" &&\
	cd templates &&\
	${GOBIN}/packr2 clean &&\
	cd .. &&\
	go mod tidy

.PHONY: tools
tools:
	GOBIN=${GOBIN} go get github.com/gobuffalo/packr/v2/packr2
	GOBIN=${GOBIN} go get github.com/onsi/ginkgo/ginkgo
	GOBIN=${GOBIN} go get github.com/onsi/gomega/...

compile-local:
	go build -ldflags "${LINKER_FLAGS}" -o ${DESTINATION} ./cmd/oam-ecs

compile-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags "${LINKER_FLAGS} ${RELEASE_BUILD_LINKER_FLAGS}" -o ${DESTINATION}.exe ./cmd/oam-ecs

compile-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "${LINKER_FLAGS} ${RELEASE_BUILD_LINKER_FLAGS}" -o ${DESTINATION}-amd64 ./cmd/oam-ecs

compile-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "${LINKER_FLAGS} ${RELEASE_BUILD_LINKER_FLAGS}" -o ${DESTINATION} ./cmd/oam-ecs

.PHONY: test
test: format packr-build compile-local run-unit-test run-e2e-test packr-clean

run-unit-test:
	go test -race -cover -count=1 -coverprofile ${COVERAGE} ${PACKAGES}

run-e2e-test:
	go test ./integ-tests...

generate-coverage: ${COVERAGE}
	go tool cover -html=${COVERAGE}

${COVERAGE}: test

.PHONY: clean
clean:
	- rm -rf ./bin
