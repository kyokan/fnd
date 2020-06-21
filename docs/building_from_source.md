# Building From Source

This guide will walk you through building `ddrpd` and `ddrpcli` from
source. Note that installation via our pre-built binaries is usually
quicker and easier unless you are working on the reference
implementation itself or are otherwise unable to use them.

## Step 1: Gather Required Software

You'll need to install the following dependencies before building either
binary:

1.  `make`.
2.  `go` version 1.12 or above.
3.  `protoc` version 3 or above.
4.  `protoc-gen-go` version 3 or above.

## Step 2: Build The Project

Both binaries can be built using the `Makefile` in the root directory of
the repository. The following `make` targets are relevant:

  - `make all`: Builds both `ddrpd` and `ddrpcli`
  - `make ddrpd`: Builds only `ddrpd`.
  - `make ddrpcli`: Builds only `ddrpcli`.
  - `make proto`: Builds only the gRPC service files. This target is
    executed as a dependency of the above targets.
  - `make install`: Builds both `ddrpd` and `ddrpcli`, then places the
    binaries in `/usr/local/bin`.
