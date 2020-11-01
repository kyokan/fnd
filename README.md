# fnd

Footnote is a protocol for synchronizing different types of messages across a peer-to-peer network. This repository contains the source code for `fnd` - the protocol's reference implementation in Go.

## Building From Source

This guide will walk you through building `fnd` from source.

### Step 1: Gather Required Software

You'll need to install the following dependencies before building either
binary:

1.  `make`.
2.  `go` version 1.12 or above.
3.  `protoc` version 3 or above.
4.  `protoc-gen-go` version 3 or above.

### Step 2: Install Dependency

1. Copy `/fnd/` directory to your `GOROOT` (e.g. `/usr/local/go/fnd`).
2. `go mod vendor` to download dependency to `/vendor/`.
3. `cp -R ./fnd/ ./vendor/fnd/` to copy source file to the vendor directory.

### Step 3: Build The Project

Both binaries can be built using the `Makefile` in the root directory of
the repository. The following `make` targets are relevant:

  - `make all`: Builds `fnd`.
  - `make proto`: Builds only the gRPC service files. This target is
    executed as a dependency of the above targets.
  - `make install`: Builds `fnd`, then places the
    binaries in `/usr/local/bin`.
    

## Starting `fnd`

Run the command `fnd init`. This will create a `.fnd` directory in
your home folder. You can customize the location of `fnd`'s home
folder by passing the `--home` flag, but bear in mind that you will have
to pass the `--home` flag with all future invocations of `fnd` if you
decide to place it in a non-standard location.

If you poke around in the `.fnd` directory, you'll see the following
things:

1.  An `identity` file, which contains the private key used to generate
    your node's identity on the network.
2.  A `db` folder, which will contain `fndd`'s local database.
3.  A `blobs` folder, which will contain raw blob data.
4.  A `config.toml` file, which contains configuration directives that
    alter `fnd`'s behavior.

You shouldn't need to change any configuration; the defaults will work
just fine.

Once `fnd` is initialized, run the command `fnd start`.


## Configuration

`fnd` is configured via the `config.toml` in its home directory. By
default, `fnd`'s home directory is set to `~/.fnd`, however this
location can be changed through the `--home` CLI flag.

Configuration options are grouped by the subsystem the control. For
example, all configuration options that affect `fnd`'s peer-to-peer
networking are grouped under the `p2p` heading. Each of these groups and
their allowed options are described below. Note that all of these values
can be left as their defaults - no configuration changes are needed for
`fnd` to run.

### Global Directives

|                   |            |           |                                                                                                                               |
| ----------------- | ---------- | --------- | ----------------------------------------------------------------------------------------------------------------------------- |
| Directive         | Type       | Default   | Description                                                                                                                   |
| `enable_profiler` | `bool`     | `false`   | Enables/disables Golang's `pprof` profiling server. The server will listen on port `9090` if enabled.                         |
| `log_level`       | `string`   | `info`    | Sets `fnd`'s log level.                                                                                                       |
| `ban_lists`       | `[]string` | Empty     | Sets Footnote's protocol-level ban lists.                                                                                     |

### Handshake Resolver Directives

|               |                |                        |                                                           |
| ------------- | -------------- | ---------------------- | --------------------------------------------------------- |
| Directive     | Type           | Default                | Description                                               |
| `host`            | `string`    | `127.0.0.1` | host to hsd node           |
| `api_key`         | `string`    | (empty)     | api_key to hsd node |
| `port`            | `number`       | `12037` | port to hsd node                |
| `base_path`       | `string`       | (empty) | base_path to hsd                  |

### Peer-To-Peer Directives

These directives control the behavior of `fnd`'s peer-to-peer
networking.

|                         |          |           |                                                                                                                                                             |
| ----------------------- | -------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Directive               | Type     | Default   | Description                                                                                                                                                 |
| `seed_peers`       | `[]string` | `[]`   | A list of bootstrap peers for your node to peer with. These should be specified as a comma-separated list of items with the format `<peer-id>@<ip>:<port>`. |
| `connection_timeout_ms` | `uint`   | `5000`    | The number of milliseconds `fnd` will wait for a new peer connection to complete.                                                                         |
| `host`                  | `string` | `0.0.0.0` | The IP address `fnd` should listen on for incoming connections.                                                                                           |
| `max_inbound_peers`     | `uint`   | `117`     | The maximum number of inbound peer connections.                                                                                                             |
| `max_outbound_peers`    | `uint`   | `8`       | The maximum number of outbound peer connections.                                                                                                            |

### RPC Directives

These directives control the behavior of `fnd`'s gRPC server, which is
used by the CLI and other clients to perform actions on the node.

|           |          |             |                                            |
| --------- | -------- | ----------- | ------------------------------------------ |
| Directive | Type     | Default     | Description                                |
| `host`    | `string` | `127.0.0.1` | The host that the server should listen on. |
| `port`    | `uint`   | `9098`      | The port that the server should listen on. |

## Daemonization

`fnd` is unopinionated regarding how it is deployed. For long-lived
deployments, however, we recommend using some kind of process management
tool such as `systemd` or `supervisor` to daemonize `fnd` and manage
its lifecycle.

A simplified version of the `systemd` unit file is included below:

    [Unit]
    Description=Footnote daemon.
    
    [Service]
    Type=simple
    ExecStart=/usr/bin/fnd start
    User=fnd
    
    [Install]
    WantedBy=multi-user.target

## Logging

`fnd` writes all log output to `stderr`. It does not include any
in-process functionality to write to or manage log files. To store
`fnd` log output, we recommend either using `syslog` or redirecting
log output to a file.

The following list outlines the allowed `log_level` configuration
directives and what each means:

  - `trace`: Outputs absolutely everything. Very verbose.
  - `debug`: Outputs almost everything. Verbose.
  - `info`: Outputs log information about core `fndd` services and
    state changes.
  - `warn`: Outputs warnings.
  - `error`: Outputs errors.
  - `fatal`: Outputs only fatal errors.

## Banning Names

If a name is hosting content that you find objectionable, or is illegal
in your jurisdiction, you can ban it from your node by creating a ban
list and setting the `ban_lists` configuration directive in your
`config.toml`. Ban lists are newline-separated lists of names starting
with `FNBAN:v1`. For example:

    FNBAN:v1
    bannedname1.
    anotherbadname.

We suggest hosting your ban lists on some server so that they can be
used across multiple nodes. For example, you could create a ban list as
a secret GitHub gist and use its `raw.githubusercontent.com` link in
your configuration.


