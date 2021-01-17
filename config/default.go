package config

import (
	"bytes"
	"io"
	"os"
	"path"
	"text/template"

	"fnd/log"

	"github.com/pkg/errors"
)

var DefaultConfig = Config{
	BanLists:       []string{},
	LogLevel:       log.LevelInfo.String(),
	EnableProfiler: false,
	Heartbeat: HeartbeatConfig{
		Moniker: "",
		URL:     "https://www.ddrpscan.com/heartbeat",
	},
	P2P: P2PConfig{
		Host: "0.0.0.0",
		DNSSeeds: []string{
			"seeds.ddrp.network",
		},
		FixedSeeds:          []string{},
		MaxInboundPeers:     117,
		MaxOutboundPeers:    8,
		ConnectionTimeoutMS: 5000,
	},
	RPC: RPCConfig{
		Host: "127.0.0.1",
		Port: 9098,
	},
	HNSResolver: HNSResolverConfig{
		Host:     "http://127.0.0.1",
		Port:     12037,
		BasePath: "",
		APIKey:   "",
	},
	Tuning: TuningConfig{
		UpdateQueue: UpdateQueueConfig{
			MaxLen:         1000,
			ReapIntervalMS: 5000,
		},
		Updater: UpdaterConfig{
			PollIntervalMS: 100,
			Workers:        2,
		},
		Syncer: SyncerConfig{
			TreeBaseResponseTimeoutMS: 10000,
			SectorResponseTimeoutMS:   15000,
		},
		SectorServer: SectorServerConfig{
			CacheExpiryMS: 5000,
		},
		PeerExchanger: PeerExchangerConfig{
			SampleSize:         12,
			RequestIntervalMS:  60 * 60 * 1000,
			MaxSentPeers:       255,
			MaxReceivedPeers:   255,
			MaxConcurrentDials: 2,
		},
		NameImporter: NameImporterConfig{
			ConfirmationDepth:     24,
			CheckIntervalMS:       60000,
			Workers:               5,
			VerificationThreshold: 0.90,
		},
		Heartbeat: HeartbeaterConfig{
			IntervalMS: 30000,
			TimeoutMS:  10000,
		},
		NameSyncer: NameSyncerConfig{
			Workers:                 2,
			SampleSize:              7,
			UpdateResponseTimeoutMS: 5000,
			IntervalMS:              60 * 60 * 1000,
			SyncResponseTimeoutMS:   60000,
		},
	},
}

const defaultConfigTemplateText = `# fnd Config File

# List of ban list URLs.
ban_lists = []

# Enables pprof profiling.
enable_profiler = {{.EnableProfiler}}

# Sets the log level. Can be one of the following values:
# - error
# - warn
# - info
# - debug
# - trace
log_level = "{{.LogLevel}}"

# Configures heartbeating, which announces this
# node's moniker and peer ID to the provided URL.
[heartbeat]
  # Sets the node's heartbeat moniker.
  moniker = "{{.Heartbeat.Moniker}}"
  # Sets the URL the node will heartbeat to.
  url = "{{.Heartbeat.URL}}"

# Configures the connection to the Handshake network. Footnote assumes
# that HSD is hosted at a url with the following format:
# <host>:<port>/<base_path>.
[hns_resolver]
  # Sets the HSD connection's API key.
  api_key = "{{.HNSResolver.APIKey}}"
  # Sets the HSD connection's base path.
  base_path = "{{.HNSResolver.BasePath}}"
  # Sets the HSD connection's host and protocol.
  host = "{{.HNSResolver.Host}}"
  # Sets the HSD connection's port.
  port = {{.HNSResolver.Port}}

# Configures the behavior of this node's peer-to-peer
# connections.
[p2p]
  # Sets how long to wait for a remote peer to respond
  # before disconnecting.
  connection_timeout_ms = {{.P2P.ConnectionTimeoutMS}}
  # Sets the set of domain names to query for seed nodes.
  # A records belonging to nodes in this list will be
  # connected to during node startup.
  dns_seeds = ["{{index .P2P.DNSSeeds 0}}"]
  # Sets the IP this node should listen on. Should be set to 0.0.0.0
  # for all Internet-accessible nodes.
  host = "{{.P2P.Host}}"
  # Sets the maximum number of inbound peers this node will handle. All
  # additional inbound peers will be rejected once this number is reached.
  # The default of 117 was chosen to match Bitcoin.
  max_inbound_peers = {{.P2P.MaxInboundPeers}}
  # Sets the maximum number of outbound peers this node will handle. The node
  # will not connect to any additional peers once this number is reached. The
  # default of 8 was chosen to match Bitcoin.
  max_outbound_peers = {{.P2P.MaxOutboundPeers}}
  # Sets a list of fixed seed peers. Items should be formatted as <peer-id>@<ip>.
  seed_peers = []

# Configures the behavior of this node's RPC server.
[rpc]
  # Sets the IP this node should listen for RPC requests on.
  # For the most part, this should be set to 127.0.0.1. Exposing
  # fnd's RPC port to the public internet is not safe.
  host = "{{.RPC.Host}}"
  # Sets the port this node should listen for RPC requests on.
  port = {{.RPC.Port}}

# Configures various internal tuning parameters. Unless directed otherwise
# or you know what you are doing, these values should be left as their
# defaults.
[tuning]

  # Configures how often fnd will perform heartbeats and
  # when to time out heartbeat requests.
  [tuning.heartbeat]
    interval_ms = {{.Tuning.Heartbeat.IntervalMS}}
    timeout_ms = {{.Tuning.Heartbeat.TimeoutMS}}

  # Configures how fnd scans the Handshake blockchain for
  # new DDRPKEY records.
  [tuning.name_importer]
    # Sets how often fnd scans for new names.
    check_interval_ms = {{.Tuning.NameImporter.CheckIntervalMS}}
    # Sets how many blocks fnd will wait before considering
    # a Handshake name record to be finalized. Changing this
    # value to something lower than the default will lead to
    # the network rejecting updates originating from this node.
    confirmation_depth = {{.Tuning.NameImporter.ConfirmationDepth}}
    # Sets how many blocks should be fetched from HSD concurrently.
    workers = {{.Tuning.NameImporter.Workers}}
    # Sets the minimum sync percentage fnd will accept from HSD before
    # importing names.
    verification_threshold = {{.Tuning.NameImporter.VerificationThreshold}}

  # Configures how fnd scans for updates in the background.
  [tuning.name_syncer]
    # Sets how often updates will be scanned for.
    interval_ms = {{.Tuning.NameSyncer.IntervalMS}}
    # Sets how many peers will be queried for updates.
    sample_size = {{.Tuning.NameSyncer.SampleSize}}
    # Sets how long the name syncer will wait for syncs to complete before proceeding
    # to the next name.
    sync_response_timeout_ms = {{.Tuning.NameSyncer.SyncResponseTimeoutMS}}
    # Sets how long the name syncer will wait for updates before proceeding to the
    # next name.
    update_response_timeout_ms = {{.Tuning.NameSyncer.UpdateResponseTimeoutMS}}
    # Sets how many names will be synchronized concurrently.
    workers = {{.Tuning.NameSyncer.Workers}}

  # Configures how fnd exchanges peers with the rest of the network.
  [tuning.peer_exchanger]
    # Sets how many concurrent dials fnd will make when it
    # receives exchanged peers.
    max_concurrent_dials = {{.Tuning.PeerExchanger.MaxConcurrentDials}}
    # Sets the maximum number of peers fnd will process after
    # receiving exchanged peers.
    max_received_peers = {{.Tuning.PeerExchanger.MaxReceivedPeers}}
    # Sets the maximum number of peers fnd will send after receiving a
    # request for peers.
    max_sent_peers = {{.Tuning.PeerExchanger.MaxSentPeers}}
    # Sets how often fnd will request new peers.
    request_interval_ms = {{.Tuning.PeerExchanger.RequestIntervalMS}}
    # Sets how many peers fnd will request new peers from during each
    # peer exchange operation.
    sample_size = {{.Tuning.PeerExchanger.SampleSize}}

  # Configures how fnd serves sector data to peers that request it.
  [tuning.sector_server]
    # Sets how often fnd will reap in-memory cached sectors.
    cache_expiry_ms = {{.Tuning.SectorServer.CacheExpiryMS}}

  # Configures how fnd synchronizes sectors with remote peers.
  [tuning.syncer]
    # Sets how long fnd will wait for remote peers to return
    # sector data before retrying.
    sector_response_timeout_ms = {{.Tuning.Syncer.SectorResponseTimeoutMS}}
    # Sets how long fnd will wait for a remote peers to return
    # tree base data before trying another peer.
    tree_base_response_timeout_ms = {{.Tuning.Syncer.TreeBaseResponseTimeoutMS}}

  # Configures how fnd enqueues blob updates.
  [tuning.update_queue]
    # Sets the maximum length of the update queue.
    max_len = {{.Tuning.UpdateQueue.MaxLen}}
    # Sets how often fnd will reap disposed of queue entries.
    reap_interval_ms = {{.Tuning.UpdateQueue.ReapIntervalMS}}

  # Configures how fnd updates blobs.
  [tuning.updater]
    # Sets how often fnd will check the update queue for new updates.
    poll_interval_ms = {{.Tuning.Updater.PollIntervalMS}}
    # Sets how many updates fnd will process concurrently.
    workers = {{.Tuning.Updater.Workers}}
`

var defaultConfigTemplate *template.Template

func GenerateDefaultConfigFile() []byte {
	buf := new(bytes.Buffer)
	if err := defaultConfigTemplate.Execute(buf, DefaultConfig); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func ReadConfigFile(homeDir string) (*Config, error) {
	f, err := os.OpenFile(path.Join(homeDir, "config.toml"), os.O_RDONLY, 0755)
	if err != nil {
		return nil, errors.Wrap(err, "error opening config file for reading")
	}
	defer f.Close()
	cfg, err := ReadConfig(f)
	if err != nil {
		return nil, errors.Wrap(err, "error reading config file")
	}
	return cfg, nil
}

func WriteDefaultConfigFile(homeDir string) error {
	f, err := os.OpenFile(path.Join(homeDir, "config.toml"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return errors.Wrap(err, "error opening config file for writing")
	}
	defer f.Close()
	rd := bytes.NewReader(GenerateDefaultConfigFile())
	if _, err := io.Copy(f, rd); err != nil {
		return errors.Wrap(err, "error writing config file")
	}
	return nil
}

func init() {
	tmpl := template.New("defaultConfig")
	t, err := tmpl.Parse(defaultConfigTemplateText)
	if err != nil {
		panic(err)
	}
	defaultConfigTemplate = t
}
