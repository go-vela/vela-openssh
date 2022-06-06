# Copyright (c) 2022 Target Brands, Inc. All rights reserved.
#
# Use of this source code is governed by the LICENSE file in this repository.

# Versions installed for OpenSSH and SSHPass binaries.
# This is the ONLY place these hardcoded versions are set.
# They're used in the Dockerfile the GitHub Actions workflow,
# the integration tests, and the static build flags for Go.
# Note: No space between the equals and the value else issues arise.
# renovate: datasource=repology depName=alpine_3_16/openssh versioning=loose
OPENSSH_VERSION=9.0_p1-r1
# renovate: datasource=repology depName=alpine_3_16/sshpass versioning=loose
SSHPASS_VERSION=1.09-r0

# check if a git tag is already set
ifndef GITHUB_TAG
	# capture the current git tag we build the application from
	GITHUB_TAG = $(shell git describe --tag --abbrev=0)
endif

# create a list of linker flags for building the golang application
LD_FLAGS = \
	-X github.com/go-vela/vela-openssh/internal/openssh.OpenSSHVersion=${OPENSSH_VERSION} \
	-X github.com/go-vela/vela-openssh/internal/openssh.SSHPassVersion=${SSHPASS_VERSION} \
	-X github.com/go-vela/vela-openssh/internal/openssh.PluginVersion=${GITHUB_TAG}

# The `clean` target is intended to clean the workspace
# and prepare the local changes for submission.
#
# Usage: `make clean`
.PHONY: clean
clean: tidy vet fmt fix

# The `run` target is intended to build and
# execute the Docker image for the plugin.
#
# Usage: `make run`
.PHONY: run
run: build docker-build docker-run

# The `tidy` target is intended to clean up
# the Go module files (go.mod & go.sum).
#
# Usage: `make tidy`
.PHONY: tidy
tidy:
	@echo
	@echo "### Tidying Go module"
	@go mod tidy

# The `vet` target is intended to inspect the
# Go source code for potential issues.
#
# Usage: `make vet`
.PHONY: vet
vet:
	@echo
	@echo "### Vetting Go code"
	@go vet ./...

# The `fmt` target is intended to format the
# Go source code to meet the language standards.
#
# Usage: `make fmt`
.PHONY: fmt
fmt:
	@echo
	@echo "### Formatting Go Code"
	@go fmt ./...

# The `fix` target is intended to rewrite the
# Go source code using old APIs.
#
# Usage: `make fix`
.PHONY: fix
fix:
	@echo
	@echo "### Fixing Go Code"
	@go fix ./...

# The `test` target is intended to run
# the tests for the Go source code.
#
# Usage: `make test`
.PHONY: test
test:
	@echo
	@echo "### Testing Go Code"
	@go test ./...

# The `test-cover` target is intended to run
# the tests for the Go source code and then
# open the test coverage report.
#
# Usage: `make test-cover`
.PHONY: test-cover
test-cover:
	@echo
	@echo "### Creating test coverage report"
	@go test -covermode=atomic -coverprofile=coverage.out ./...
	@echo
	@echo "### Opening test coverage report"
	@go tool cover -html=coverage.out

# The `build` target is intended to compile
# the Go source code into a binary.
#
# Usage: `make build`
.PHONY: build
build:
	@echo
	@echo "### Building release/vela-scp binary"
	GOOS=linux CGO_ENABLED=0 \
		go build -a \
		-ldflags '${LD_FLAGS}' \
		-o release/vela-scp \
		github.com/go-vela/vela-openssh/cmd/vela-scp
	@echo
	@echo "### Building release/vela-ssh binary"
	GOOS=linux CGO_ENABLED=0 \
		go build -a \
		-ldflags '${LD_FLAGS}' \
		-o release/vela-ssh \
		github.com/go-vela/vela-openssh/cmd/vela-ssh

# The `build-static` target is intended to compile
# the Go source code into a statically linked binary.
#
# Usage: `make build-static`
.PHONY: build-static
build-static:
	@echo
	@echo "### Building static release/vela-scp binary"
	GOOS=linux CGO_ENABLED=0 \
		go build -a \
		-ldflags '-s -w -extldflags "-static" ${LD_FLAGS}' \
		-o release/vela-scp \
		github.com/go-vela/vela-openssh/cmd/vela-scp
	@echo
	@echo "### Building static release/vela-ssh binary"
	GOOS=linux CGO_ENABLED=0 \
		go build -a \
		-ldflags '-s -w -extldflags "-static" ${LD_FLAGS}' \
		-o release/vela-ssh \
		github.com/go-vela/vela-openssh/cmd/vela-ssh

# The `build-static-ci` target is intended to compile
# the Go source code into a statically linked binary
# when used within a CI environment.
#
# Usage: `make build-static-ci`
.PHONY: build-static-ci
build-static-ci:
	@echo
	@echo "### Building CI static release/vela-scp binary"
	@go build -a \
		-ldflags '-s -w -extldflags "-static" ${LD_FLAGS}' \
		-o release/vela-scp \
		github.com/go-vela/vela-openssh/cmd/vela-scp
	@echo
	@echo "### Building CI static release/vela-ssh binary"
	@go build -a \
		-ldflags '-s -w -extldflags "-static" ${LD_FLAGS}' \
		-o release/vela-ssh \
		github.com/go-vela/vela-openssh/cmd/vela-ssh

# The `check` target is intended to output all
# dependencies from the Go module that need updates.
#
# Usage: `make check`
.PHONY: check
check: check-install
	@echo
	@echo "### Checking dependencies for updates"
	@go list -u -m -json all | go-mod-outdated -update

# The `check-direct` target is intended to output direct
# dependencies from the Go module that need updates.
#
# Usage: `make check-direct`
.PHONY: check-direct
check-direct: check-install
	@echo
	@echo "### Checking direct dependencies for updates"
	@go list -u -m -json all | go-mod-outdated -direct

# The `check-full` target is intended to output
# all dependencies from the Go module.
#
# Usage: `make check-full`
.PHONY: check-full
check-full: check-install
	@echo
	@echo "### Checking all dependencies for updates"
	@go list -u -m -json all | go-mod-outdated

# The `check-install` target is intended to download
# the tool used to check dependencies from the Go module.
#
# Usage: `make check-install`
.PHONY: check-install
check-install:
	@echo
	@echo "### Installing psampaz/go-mod-outdated"
	@go get -u github.com/psampaz/go-mod-outdated

# The `bump-deps` target is intended to upgrade
# non-test dependencies for the Go module.
#
# Usage: `make bump-deps`
.PHONY: bump-deps
bump-deps: check
	@echo
	@echo "### Upgrading dependencies"
	@go get -u ./...

# The `bump-deps-full` target is intended to upgrade
# all dependencies for the Go module.
#
# Usage: `make bump-deps-full`
.PHONY: bump-deps-full
bump-deps-full: check
	@echo
	@echo "### Upgrading all dependencies"
	@go get -t -u ./...

# The `docker-build` target is intended to build
# the Docker image for the plugin.
#
# Usage: `make docker-build`
.PHONY: docker-build
docker-build:
	@echo
	@echo "### Building vela-scp:local image"
	@docker build -f Dockerfile.scp --no-cache --build-arg OPENSSH_VERSION=${OPENSSH_VERSION} --build-arg SSHPASS_VERSION=${SSHPASS_VERSION} -t vela-scp:local .
	@docker build -f Dockerfile.ssh --no-cache --build-arg OPENSSH_VERSION=${OPENSSH_VERSION} --build-arg SSHPASS_VERSION=${SSHPASS_VERSION} -t vela-ssh:local .

# The `docker-test` target is intended to execute
# the Docker image for the plugin with test variables
# for integration testing the plugins work with other systems.
#
# Usage: `make docker-test`
.PHONY: docker-test
docker-test:
	@test/integration-tests.sh

# The `docker-run-scp` target is intended to execute
# the Docker image for the scp plugin.
#
# Usage: `make docker-run-scp`
.PHONY: docker-run-scp
docker-run-scp:
	@echo
	@echo "### Executing vela-scp:local image"
	@docker run --rm \
		-e PARAMETER_SOURCE \
		-e PARAMETER_TARGET \
		-e PARAMETER_IDENTITY_FILE_PATH \
		-e PARAMETER_IDENTITY_FILE_CONTENTS \
		-e PARAMETER_SCP_FLAG \
		-e PARAMETER_SSHPASS_PASSWORD \
		-e PARAMETER_SSHPASS_PASSPHRASE \
		-e PARAMETER_SSHPASS_FLAG \
		-e PARAMETER_CI \
		-v $(shell pwd):/home \
		vela-scp:local

# The `docker-run-ssh` target is intended to execute
# the Docker image for the ssh plugin.
#
# Usage: `make docker-run-ssh`
.PHONY: docker-run-ssh
docker-run-ssh:
	@echo
	@echo "### Executing vela-ssh:local image"
	@docker run --rm \
		-e PARAMETER_DESTINATION \
		-e PARAMETER_COMMAND \
		-e PARAMETER_IDENTITY_FILE_PATH \
		-e PARAMETER_IDENTITY_FILE_CONTENTS \
		-e PARAMETER_SSH_FLAG \
		-e PARAMETER_SSHPASS_PASSWORD \
		-e PARAMETER_SSHPASS_PASSPHRASE \
		-e PARAMETER_SSHPASS_FLAG \
		-e PARAMETER_CI \
		-v $(shell pwd):/home \
		vela-scp:local
