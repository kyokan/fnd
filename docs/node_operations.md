# Node Operations

## Daemonization

`ddrpd` is unopinionated regarding how it is deployed. For long-lived
deployments, however, we recommend using some kind of process management
tool such as `systemd` or `supervisor` to daemonize `ddrpd` and manage
its lifecycle.

A simplified version of the `systemd` unit file Kyokan uses for its seed
nodes is included below:

    [Unit]
    Description=DDRPD daemon.
    
    [Service]
    Type=simple
    ExecStart=/usr/bin/ddrpd start
    User=ddrpd
    
    [Install]
    WantedBy=multi-user.target

## Logging

`ddrpd` writes all log output to `stderr`. It does not include any
in-process functionality to write to or manage log files. To store
`ddrpd` log output, we recommend either using `syslog` or redirecting
log output to a file.

The following list outlines the allowed `log_level` configuration
directives and what each means:

  - `trace`: Outputs absolutely everything. Very verbose.
  - `debug`: Outputs almost everything. Verbose.
  - `info`: Outputs log information about core `ddrpd` services and
    state changes.
  - `warn`: Outputs warnings.
  - `error`: Outputs errors.
  - `fatal`: Outputs only fatal errors.

## Banning Names

If a name is hosting content that you find objectionable, or is illegal
in your jurisdiction, you can ban it from your node by creating a ban
list and setting the `ban_lists` configuration directive in your
`config.toml`. Ban lists are newline-separated lists of names starting
with `DDRPBAN:v1`. For example:

    DDRPBAN:v1
    bannedname1.
    anotherbadname.

We suggest hosting your ban lists on some server so that they can be
used across multiple nodes. For example, you could create a ban list as
a secret GitHub gist and use its `raw.githubusercontent.com` link in
your configuration.

See [PIP-6](./spec/pip-6.html) for more information about creating ban
lists.
