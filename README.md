# Go Build Action [![Heimdall](https://heimdall.hashicorp.services/api/v1/assets/actions-go-build/badge.svg?key=5c34743984c6ac17fabc3e68b7f6d34620de4e877ab1c529405ed7f4843147bf)](https://heimdall.hashicorp.services/site/assets/actions-go-build) [![CI](https://github.com/hashicorp/actions-go-build/actions/workflows/test.yml/badge.svg)](https://github.com/hashicorp/actions-go-build/actions/workflows/test.yml)

_**Build and package a (reproducible) Go binary.**_

- **Define** the build.
- **Assert** that it is reproducible (optionally).
- **Use** the resultant artifacts in your workflow.

_This is intended for internal HashiCorp use only; Internal folks please refer to RFC **ENGSRV-084** for more details._

<!-- insert:dev/docs/table_of_contents -->
* [Features](#features)
* [Local Usage](#local-usage)
* [Usage in GHA](#usage-in-gha)
  * [Examples](#examples)
  * [Inputs](#inputs)
  * [Build Instructions](#build-instructions)
  * [Ensuring Reproducibility](#ensuring-reproducibility)
* [Development](#development)
<!-- end:insert:dev/docs/table_of_contents -->

## Features

- **Results are zipped** using standard HashiCorp naming conventions.
- **You can include additional files** in the zip like licenses etc.
- **Convention over configuration** means minimal config required.
- **Reproducibility** is checked at build time.
- **Fast feedback** if accidental nondeterminism is introduced.

## Local Usage

The core functionality of this action is contained in a Go CLI, which
you can also install and use locally. See [the CLI docs](docs/cli.md)
for more.

## Usage in GHA

This Action can run on both Ubuntu and macOS runners.

### Examples

#### Minimal(ish) Example

This example shows building a single `linux/amd64` binary.

[See this simple example workflow running here](https://github.com/hashicorp/actions-reproducible-build/actions/workflows/example.yml).

<!-- insert:dev/docs/print_example_workflow example.yml -->
```yaml
name: Minimal Example (main)
on: push
jobs:
  example:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Build
        uses: hashicorp/actions-go-build@main
        with:
          go_version: 1.18
          os: linux
          arch: amd64
          work_dir: testdata/example-app
          debug: true
          instructions: |
            go build -o "$BIN_PATH" -trimpath -buildvcs=false
```
<!-- end:insert:dev/docs/print_example_workflow example.yml -->

#### More Realistic Example

This example shows usage of the action inside a matrix configured to produce
binaries for different platforms. It also injects the version, revision, and
revision time into the binary via `-ldflags`, uses the `netcgo` tag for darwin,
and disables CGO for linux and windows builds.

[See this matrix example workflow running here](https://github.com/hashicorp/actions-reproducible-build/actions/workflows/example-matrix.yml).

<!-- insert:dev/docs/print_example_workflow example-matrix.yml -->
```yaml
name: Matrix Example (main)
on: push
jobs:
  example:
    runs-on: ${{ matrix.runner }}
    strategy:
      matrix:
        include:
          - { runner: macos-latest,  os: darwin,  arch: amd64, tags: netcgo        }
          - { runner: macos-latest,  os: darwin,  arch: arm64, tags: netcgo        }
          - { runner: ubuntu-latest, os: linux,   arch: amd64, env:  CGO_ENABLED=0 }
          - { runner: ubuntu-latest, os: linux,   arch: amd64, env:  CGO_ENABLED=0 }
          - { runner: ubuntu-latest, os: windows, arch: amd64, env:  CGO_ENABLED=0 }
    steps:
      - uses: actions/checkout@v3
      - name: Build
        uses: hashicorp/actions-go-build@main
        with:
          product_name: example-app
          product_version: 1.2.3
          go_version: 1.18
          os: ${{ matrix.os }}
          arch: ${{ matrix.arch }}
          instructions: |-
            cd ./testdata/example-app && \
            ${{ matrix.env }} \
              go build \
                -o "$BIN_PATH" \
                -trimpath \
                -buildvcs=false \
                -tags="${{ matrix.tags }}" \
                -ldflags "
                  -X 'main.Version=$PRODUCT_VERSION'
                  -X 'main.Revision=$PRODUCT_REVISION'
                  -X 'main.RevisionTime=$PRODUCT_REVISION_TIME'
                "
```
<!-- end:insert:dev/docs/print_example_workflow example-matrix.yml -->

### Inputs

<!-- insert:dev/docs/inputs_doc -->
|  Name                                      |  Description                                                                                              |
|  -----                                     |  -----                                                                                                    |
|  `product_name`&nbsp;_(optional)_          |  Used to calculate default `bin_name` and `zip_name`. Defaults to repository name.                        |
|  `product_version`&nbsp;_(optional)_       |  Full version of the product being built (including metadata).                                            |
|  `product_version_meta`&nbsp;_(optional)_  |  The metadata field of the version.                                                                       |
|  **`go_version`**&nbsp;_(required)_        |  Version of Go to use for this build.                                                                     |
|  **`os`**&nbsp;_(required)_                |  Target product operating system.                                                                         |
|  **`arch`**&nbsp;_(required)_              |  Target product architecture.                                                                             |
|  `reproducible`&nbsp;_(optional)_          |  Assert that this build is reproducible. Options are `assert` (the default), `report`, or `nope`.         |
|  `bin_name`&nbsp;_(optional)_              |  Name of the product binary generated. Defaults to `product_name` minus any `-enterprise` suffix.         |
|  `zip_name`&nbsp;_(optional)_              |  Name of the product zip file. Defaults to `<product_name>_<product_version>_<os>_<arch>.zip`.            |
|  `work_dir`&nbsp;_(optional)_              |  The working directory, to run the instructions in. Defaults to the current directory.                    |
|  **`instructions`**&nbsp;_(required)_      |  Build instructions to generate the binary. See [Build Instructions](#build-instructions) for more info.  |
|  `debug`&nbsp;_(optional)_                 |  Enable debug-level logging.                                                                              |
|  `clean`&nbsp;_(optional)_                 |  Build with the clean flag on.       |
<!-- end:insert:dev/docs/inputs_doc -->

### Build Instructions

The `instructions` input is a bash script that builds the product binary.
It should be kept as simple as possible.
Typically this will be a simple `go build` invocation,
but it could hit a make target, or call another script.
See [Example Build Instructions](#example-build-instructions)
below for examples of valid instructions.

The instructions _must_ use the environment variable `$BIN_PATH`
because the minimal thing they can do is to write the compiled binary to `$BIN_PATH`.

In order to add other files like licenses etc to the zip file, you need to
write them into `$TARGET_DIR` in your build instructions.

#### Build Environment

When the `instructions` are executed, there are a set of environment variables
already exported that you can make use of
(see [Environment Variables](#environment-variables) below).

#### Environment Variables

<!-- insert:dev/docs/environment_doc -->
|  Name                     |  Description                                                                    |
|  -----                    |  -----                                                                          |
|  `PRODUCT_NAME`           |  Same as the `product_name` input.                                              |
|  `PRODUCT_VERSION`        |  Same as the `product_version` input.                                           |
|  `PRODUCT_REVISION`       |  The git commit SHA of the product repo being built.                            |
|  `PRODUCT_REVISION_TIME`  |  UTC timestamp of the `PRODUCT_REVISION` commit in iso-8601 format.             |
|  `OS`                     |  Same as the `os` input.                                                        |
|  `ARCH`                   |  Same as the `arch` input.                                                      |
|  `GOOS`                   |  Same as `OS`.                                                                  |
|  `GOARCH`                 |  Same as `ARCH`.                                                                |
|  `WORKTREE_DIRTY`         |  Whether the workrtree is dirty (`true` or `false`).                            |
|  `WORKTREE_HASH`          |  Unique hash of the work tree. Same as PRODUCT_REVISION unless WORKTREE_DIRTY.  |
|  `TARGET_DIR`             |  Absolute path to the zip contents directory.                                   |
|  `BIN_PATH`               |  Absolute path to where instructions must write Go executable.                  |
<!-- end:insert:dev/docs/environment_doc -->

#### Reproducibility Assertions

The `reproducible` input has three options:

- `assert` (the default) means perform a verification build and fail if it's not identical to the primary build.
- `report` means perform a verification build, log the results, but do not fail.
- `nope`   means do not perform a verification build at all.

See [Ensuring Reproducibility](#ensuring-reproducibility), below for tips on making your build reproducible.

#### Example Build Instructions

The examples below all illustrate valid build instructions using `go build` flags
that give the build some chance at being reproducible.

---

Simplest Go 1.17 invocation. (Uses `-trimpath` to aid with reproducibility.)

```yaml
instructions: go build -o "$BIN_PATH" -trimpath
```

---

Simplest Go 1.18+ invocation. (Additionally uses `-buildvcs=false` to aid with reproducibility.)

```yaml
instructions: go build -o "$BIN_PATH" -trimpath -buildvcs=false
```

---

More complex build, including copying a license file into the zip and `cd`ing into
a subdirectory to perform the go build.

```yaml
instructions: |
	cp LICENSE "$TARGET_DIR/"
	cd sub/directory
	go build -o "$BIN_PATH" -trimpath -buildvcs=false
```

---

An example using `make`:

```yaml
instructions: make build
```

With this Makefile:

```Makefile
build:
	go build -o "$BIN_PATH" -trimpath -buildvcs=false
```

---

See also the [example workflow](#example-workflow) above,
which injects info into the binary using `-ldflags`.


### Ensuring Reproducibility

If you are aiming to create a reproducible build, you need to at a minimum ensure that
your build is independent from the _time_ it is run, and from the _path_ that the module
is at on the filesystem.

#### Build Time

Embedding the actual 'build time' into your binary will ensure that it isn't reproducible,
because this time will be different for each build. Instead, you can use the
`PRODUCT_REVISION_TIME` which is the time of the latest commit, which will be the same
for each build of that commit.

#### Build Path

By default `go build` embeds the absolute path to the source files inside the binaries
for use in stack traces and debugging. However, this reduces reproducibility because
that path is likely to be different for different builds.

Use the `-trimpath` flag to remove the portion of the path that is dependent on the
absolute module path to aid with reproducibility.

#### VCS information

Go 1.18+ embeds information about the current checkout directory of your code, including
modified and new files. In some cases this interferes with reproducibility. You can
turn this off using the `-buildvcs=false` flag.

## Development

Development docs have moved to [docs/development.md](docs/development.md).

The core functionality of this action is contained in a Go CLI, which can also be installed
and run locally. See [the CLI docs](docs/cli.md) for instructions on installing and using
the CLI.
