# Node Operations

## Daemonization

`fnd` is unopinionated regarding how it is deployed. For long-lived
deployments, however, we recommend using some kind of process management
tool such as `systemd` or `supervisor` to daemonize `fnd` and manage
its lifecycle.

A simplified version of the `systemd` unit file Kyokan uses for its seed
nodes is included below:

    [Unit]
    Description=fnd daemon.
    
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
  - `info`: Outputs log information about core `fnd` services and
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

See [PIP-6](./spec/pip-6.html) for more information about creating ban
lists.
