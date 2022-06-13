SHELL := /usr/bin/env bash -euo pipefail -c

default: test

BATS := bats -j 10 -T

test: test/bats

test/bats:
	# Running bats tests in scripts/
	@$(BATS) scripts/

.PHONY: docs
docs:
	@./dev/docs/readme_update
	@./dev/docs/changelog_update

.PHONY: debug/docs
debug/docs: export DEBUG := 1
debug/docs: docs

LDFLAGS += -X 'main.Version=1.2.3'
LDFLAGS += -X 'main.Revision=cabba9e'
LDFLAGS += -X 'main.RevisionTime=2022-05-30T14:45:00+00:00'

.PHONY: example-app
example-app:
	@cd testdata/example-app && go build -ldflags "$(LDFLAGS)" . && ./example-app

GO_BUILD := go build -trimpath -buildvcs=false -ldflags "$(LDFLAGS)" -o "$$BIN_PATH"

example: export REPRODUCIBLE             := assert
example: export INSTRUCTIONS             := cd testdata/example-app && $(GO_BUILD)
example: export OS                       := $(shell go env GOOS)
example: export ARCH                     := $(shell go env GOARCH)
example: export PRODUCT_VERSION          := 1.2.3
example: export PRODUCT_NAME             := example-app
example: export EXAMPLE_TMP              := $(shell mktemp -d)
example: export GITHUB_ENV               := $(EXAMPLE_TMP)/github_env
example: export GITHUB_STEP_SUMMARY      := $(EXAMPLE_TMP)/github_step_summary
example: export PRIMARY_BUILD_ROOT       := $(EXAMPLE_TMP)/primary
example: export VERIFICATION_BUILD_ROOT  := $(EXAMPLE_TMP)/verification
example:
	@rm -rf "$(EXAMPLE_TMP)" && mkdir -p "$(EXAMPLE_TMP)"
	@cp -rf . "$(PRIMARY_BUILD_ROOT)"
	@cd $(PRIMARY_BUILD_ROOT) && \
		source scripts/inputs.bash && \
		digest_inputs && \
		./scripts/primary_build && \
		./scripts/local_verification_build && \
		trap 'cat $(GITHUB_STEP_SUMMARY)' EXIT && \
		./scripts/compare_digests


# 'make tools' will use the brew target if on Darwin.
# Otherwise it just prints a message about dependencies.
ifeq ($(shell uname),Darwin)
tools: tools/mac/brew
else
tools:
	@echo "Please ensure that BATS, coreutils, util-linux, github-markdown-toc, and GNU parallel are installed."
endif

# tools/mac/brew tries to install dependencies on mac using homebrew.
tools/mac/brew:
	brew bundle --no-upgrade	

.PHONY: release
release:
	./dev/release/create
