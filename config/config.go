package config

import (
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"io"
	"time"
)

type Config struct {
	LogLevel       string            `mapstructure:"log_level"`
	EnableProfiler bool              `mapstructure:"enable_profiler"`
	Heartbeat      HeartbeatConfig   `mapstructure:"heartbeat"`
	P2P            P2PConfig         `mapstructure:"p2p"`
	RPC            RPCConfig         `mapstructure:"rpc"`
	HNSResolver    HNSResolverConfig `mapstructure:"hns_resolver"`
	BanLists       []string          `mapstructure:"ban_lists"`
	Tuning         TuningConfig      `mapstructure:"tuning"`
}

type HeartbeatConfig struct {
	Moniker string `mapstructure:"moniker"`
	URL     string `mapstructure:"url"`
}

type P2PConfig struct {
	Host                string   `mapstructure:"host"`
	DNSSeeds            []string `mapstructure:"dns_seeds"`
	FixedSeeds          []string `mapstructure:"seed_peers"`
	MaxInboundPeers     int      `mapstructure:"max_inbound_peers"`
	MaxOutboundPeers    int      `mapstructure:"max_outbound_peers"`
	ConnectionTimeoutMS int      `mapstructure:"connection_timeout_ms"`
}

type RPCConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type HNSResolverConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	BasePath string `mapstructure:"base_path"`
	APIKey   string `mapstructure:"api_key"`
}

type TuningConfig struct {
	Timebank      TimebankConfig      `mapstructure:"timebank"`
	UpdateQueue   UpdateQueueConfig   `mapstructure:"update_queue"`
	Updater       UpdaterConfig       `mapstructure:"updater"`
	Syncer        SyncerConfig        `mapstructure:"syncer"`
	SectorServer  SectorServerConfig  `mapstructure:"sector_server"`
	PeerExchanger PeerExchangerConfig `mapstructure:"peer_exchanger"`
	NameImporter  NameImporterConfig  `mapstructure:"name_importer"`
	Heartbeat     HeartbeaterConfig   `mapstructure:"heartbeat"`
	NameSyncer    NameSyncerConfig    `mapstructure:"name_syncer"`
}

type TimebankConfig struct {
	PeriodMS             int `mapstructure:"period_ms"`
	MinUpdateIntervalMS  int `mapstructure:"min_update_interval_ms"`
	FullUpdatesPerPeriod int `mapstructure:"full_updates_per_period"`
}

type UpdateQueueConfig struct {
	MaxLen         int `mapstructure:"max_len"`
	ReapIntervalMS int `mapstructure:"reap_interval_ms"`
}

type UpdaterConfig struct {
	PollIntervalMS int `mapstructure:"poll_interval_ms"`
	Workers        int `mapstructure:"workers"`
}

type SyncerConfig struct {
	TreeBaseResponseTimeoutMS int `mapstructure:"tree_base_response_timeout_ms"`
	SectorResponseTimeoutMS   int `mapstructure:"sector_response_timeout_ms"`
}

type SectorServerConfig struct {
	CacheExpiryMS int `mapstructure:"cache_expiry_ms"`
}

type PeerExchangerConfig struct {
	SampleSize         int `mapstructure:"sample_size"`
	ResponseTimeoutMS  int `mapstructure:"response_timeout_ms"`
	RequestIntervalMS  int `mapstructure:"request_interval_ms"`
	MaxSentPeers       int `mapstructure:"max_sent_peers"`
	MaxReceivedPeers   int `mapstructure:"max_received_peers"`
	MaxConcurrentDials int `mapstructure:"max_concurrent_dials"`
}

type NameImporterConfig struct {
	ConfirmationDepth     int     `mapstructure:"confirmation_depth"`
	CheckIntervalMS       int     `mapstructure:"check_interval_ms"`
	Workers               int     `mapstructure:"workers"`
	VerificationThreshold float64 `mapstructure:"verification_threshold"`
}

type HeartbeaterConfig struct {
	IntervalMS int `mapstructure:"interval_ms"`
	TimeoutMS  int `mapstructure:"timeout_ms"`
}

type NameSyncerConfig struct {
	Workers                 int `mapstructure:"workers"`
	SampleSize              int `mapstructure:"sample_size"`
	UpdateResponseTimeoutMS int `mapstructure:"update_response_timeout_ms"`
	IntervalMS              int `mapstructure:"interval_ms"`
	SyncResponseTimeoutMS   int `mapstructure:"sync_response_timeout_ms"`
}

func ReadConfig(r io.Reader) (*Config, error) {
	decoder := toml.NewDecoder(r)
	decoder.SetTagName("mapstructure")
	config := &Config{}
	if err := decoder.Decode(config); err != nil {
		return nil, errors.Wrap(err, "error decoding config file")
	}
	return config, nil
}

func ConvertDuration(base int, unit time.Duration) time.Duration {
	return time.Duration(base) * unit
}
