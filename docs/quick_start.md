# Quick Start

This tutorial will walk you through how to get started running `ddrpd`,
the Golang reference implementation of the DDRP protocol. While the
commands included below assume that you're running OSX or Linux, they
shouldn't be very different for Windows machines.

## Step 1: Install `hsd`

DDRP uses Handshake TLDs to address blobs, so you'll need to install
`hsd` before installing DDRP. `hsd` is Handshake's official full node
implementation.

To install `hsd`, first ensure that you have NPM and Node.js installed
on your system. If you don't, head on over to the [Node.js
Downloads](https://nodejs.org/en/download/) site and follow the
instructions there. Then, run the following commands in a long-lived
terminal:

``` bash
npm install -g hsd
hsd --index-tx
```

This will install and start a local `hsd` node for use with DDRP. You
should see logs streaming into your console as `hsd` synchronizes with
the Handshake network.

## Step 2: Get the Binaries

You can find precompiled binaries for your system and their PGP
signature on our [GitHub
Releases](https://github.com/kyokan/ddrp/releases) page. Download the
`.tgz` archive and PGP signature for your system, then verify the
archive using the command below:

``` bash
gpg --verify ddrp-<os>-<arch>.tgz.sig ddrp-<os>-<arch>.tgz
```

You should see output that looks like the following. Make sure to verfiy
that the outputted key ID is `D4B604F1`:

    gpg: Signature made Tue Jan  7 19:29:04 2020 PST using RSA key ID D4B604F1
    gpg: Good signature from "Kyokan Security <security@kyokan.io>"

If everything checks out, it's time to extract the archive onto your
`PATH`. To do this, run the following commands (note that they assume
`/usr/local/bin` is on your `$PATH`, if it isn't substitute it for a
directory that is):

``` bash
tar -C /usr/local/bin -xzvf ddrp-<os>-<arch>.tgz
```

You should now be able to run `ddrpd`, the DDRP node software, and
interact with it via `ddrpcli`, the node's command line harness.

## Step 3: Initialize `ddrpd`

Run the command `ddrpd init`. This will create a `.ddrpd` directory in
your home folder. You can customize the location of `ddrpd`'s home
folder by passing the `--home` flag, but bear in mind that you will have
to pass the `--home` flag with all future invocations of `ddrpd` if you
decide to place it in a non-standard location.

If you poke around in the `.ddrpd` directory, you'll see the following
things:

1.  An `identity` file, which contains the private key used to generate
    your node's identity on the network.
2.  A `db` folder, which will contain `ddrpd`'s local database.
3.  A `blobs` folder, which will contain raw blob data.
4.  A `config.toml` file, which contains configuration directives that
    alter `ddrpd`'s behavior.

You shouldn't need to change any configuration; the defaults will work
just fine.

If you're interested in learning more about the inner workings of
`ddrpd`, check out the documentation under the "Architecture" heading in
the left-hand panel.

## Step 4: Start `ddrpd`

In a terminal window that you're comfortable leaving open, run `ddrpd
start`. This will start `ddrpd`. You should start to see logs streaming
by in your console.

Your node will connect to a default set of bootstrap peers. You can see
the IP addresses and peer IDs of these nodes in `config.toml`. While the
bootstrap peers will provide additional peers for your node to connect
to, you can manually connect to peers using the `ddrpcli net add-peer`
command as described below:

``` bash
ddrpcli add-peer <peer-id>@<ip>:9097
```

Your node will store and automatically reconnect to any peers you
manually add. To find additional peers to manually connect to, check out
[DDRPScan](https://www.ddrpscan.com).

To view the peers your node is connected to, run `ddrpcli net
peer-info`.

## Step 5: Update Your Handshake Name

To start posting data to a DDRP blob, you'll need a Handshake TLD. You
can get one through Handshake's [Name
Auctions](https://hsd-dev.org/guides/auctions.html), using either a
custodial exchange/registrar like [Namebase](https://www.namebase.io) or
a non-custodial wallet like [Bob
Wallet](https://github.com/kyokan/bob-wallet). Namebase has an
[FAQ](https://www.namebase.io/faq/) which outlines their process.

Once you've acquired a TLD, you'll need to add a `TXT` record to its DNS
settings with the following format: `DDRPKEY:<hex-encoded public key>`.
The easiest way to generate a key pair for use with DDRP is via
`ddrpcli`. The following commands will do this:

``` bash
ddrpcli init # note: you'll only need to do this once
ddrpcli identity
```

The output of `ddrpcli identity` is the hex-encoded public key you need
to put in your TLD's `TXT` record.

## Step 6: Wait for `ddrpd` to Sync

Now that you've updated your name with the `TXT` record, `ddrpd` will
automatically discover it after 32 Handshake blocks (i.e., about 8
hours). You'll see an entry in your logs letting you know once this
happens.

## Step 7: Read/Write Your Blob

At this point, you're all set. You can run `ddrpcli blob write <your
tld>` to write data to your TLD's blob, and run `ddrpcli blob read <your
tld>` to read the data stored within it. All updates will be gossipped
around the network.

## Next Steps

Now that your node is set up and you have access to a blob, there's lots
more to do\! Why not check out the [Protocol Specification](https://github.com/ddrp-org/PIPs) to
learn more about how DDRP works, or the [RPC documentation](./rpc.md) to
learn how to integrate your own applications with DDRP.
