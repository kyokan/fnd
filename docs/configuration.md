# Configuration

`ddrpd` is configured via the `config.toml` in its home directory. By
default, `ddrpd`'s home directory is set to `~/.ddrpd`, however this
location can be changed through the `--home` CLI flag.

Configuration options are grouped by the subsystem the control. For
example, all configuration options that affect `ddrpd`'s peer-to-peer
networking are grouped under the `p2p` heading. Each of these groups and
their allowed options are described below. Note that all of these values
can be left as their defaults - no configuration changes are needed for
`ddrpd` to run.

## Global Directives

|                   |            |           |                                                                                                                               |
| ----------------- | ---------- | --------- | ----------------------------------------------------------------------------------------------------------------------------- |
| Directive         | Type       | Default   | Description                                                                                                                   |
| `enable_profiler` | `bool`     | `false`   | Enables/disables Golang's `pprof` profiling server. The server will listen on port `9090` if enabled.                         |
| `log_level`       | `string`   | `info`    | Sets `ddrpd`'s log level. See the dedicated [Logging](deployment.html) document for more information on available log levels. |
| `network`         | `string`   | `testnet` | Sets the network `ddrpd` is supposed to connect to. Can be `testnet`, `mainnet`, or `simnet`.                                 |
| `ban_lists`       | `[]string` | Empty     | Sets DDRP's protocol-level ban lists. See [Banning Names](./deployment.html#banning-names) for more info.                     |

## Resolver Directives

For more information on DDRP name resolvers and what they do, check out
the Resolvers document.

|               |                |                        |                                                           |
| ------------- | -------------- | ---------------------- | --------------------------------------------------------- |
| Group: \`\`ce | ntralized\_res | olver\`\`              |                                                           |
| Directive     | Type           | Default                | Description                                               |
| `host`        | `string`       | `192.241.221.138:8080` | Sets the location of a DDRP centralized resolver service. |

|                   |             |             |                                                          |
| ----------------- | ----------- | ----------- | -------------------------------------------------------- |
| Group: \`\`hns\_r | esolver\`\` |             |                                                          |
| Directive         | Type        | Default     | Description                                              |
| `host`            | `string`    | `127.0.0.1` | Sets the hostname of the resolving hsd node.             |
| `api_key`         | `string`    | (empty)     | Sets the API key used to authenticate with the hsd node. |
| `network`         | `string`    | `main`      | Sets the active Handshake network to pull data from.     |

## Heartbeat Directives

These directives configure node heartbeating, an optional feature of
`ddrpd` that allows you to publicly announce the existence of your node
and opt-in to global network metrics and dashboards.

By default, your node will not send heartbeats.

|           |          |         |                                                         |
| --------- | -------- | ------- | ------------------------------------------------------- |
| Directive | Type     | Default | Description                                             |
| `moniker` | `string` | (empty) | A name for your node. Will appear on public dashboards. |
| `url`     | `string` | (empty) | The server you would like your node to heartbeat to.    |

## Peer-To-Peer Directives

These directives control the behavior of `ddrpd`'s peer-to-peer
networking.

|                         |          |           |                                                                                                                                                             |
| ----------------------- | -------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Directive               | Type     | Default   | Description                                                                                                                                                 |
| `bootstrap_peers`       | `string` | (empty)   | A list of bootstrap peers for your node to peer with. These should be specified as a comma-separated list of items with the format `<peer-id>@<ip>:<port>`. |
| `connection_timeout_ms` | `uint`   | `5000`    | The number of milliseconds `ddrpd` will wait for a new peer connection to complete.                                                                         |
| `host`                  | `string` | `0.0.0.0` | The IP address `ddrpd` should listen on for incoming connections.                                                                                           |
| `max_inbound_peers`     | `uint`   | `117`     | The maximum number of inbound peer connections.                                                                                                             |
| `max_outbound_peers`    | `uint`   | `8`       | The maximum number of outbound peer connections.                                                                                                            |
| `port`                  | `uint`   | `9097`    | The port `ddrpd` should listen on for incoming connections.                                                                                                 |

## RPC Directives

These directives control the behavior of `ddrpd`'s gRPC server, which is
used by the CLI and other clients to perform actions on the node.

|           |          |             |                                            |
| --------- | -------- | ----------- | ------------------------------------------ |
| Directive | Type     | Default     | Description                                |
| `host`    | `string` | `127.0.0.1` | The host that the server should listen on. |
| `port`    | `uint`   | `9098`      | The port that the server should listen on. |
