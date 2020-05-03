package cmd

import (
	"fmt"
	"github.com/ddrp-org/ddrp/blob"
	"github.com/ddrp-org/ddrp/cli"
	"github.com/ddrp-org/ddrp/config"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/ddrp-org/ddrp/log"
	"github.com/ddrp-org/ddrp/p2p"
	"github.com/ddrp-org/ddrp/protocol"
	"github.com/ddrp-org/ddrp/rpc"
	"github.com/ddrp-org/ddrp/service"
	"github.com/ddrp-org/ddrp/store"
	"github.com/ddrp-org/ddrp/util"
	"github.com/ddrp-org/ddrp/version"
	"github.com/mslipper/handshake/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the daemon.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfigFile(configuredHomeDir)
		if err != nil {
			return errors.Wrap(err, "error reading config file")
		}
		rpcHost := cfg.RPC.Host
		rpcPort := cfg.RPC.Port
		p2pHost := cfg.P2P.Host
		logLevel, err := log.NewLevel(cfg.LogLevel)
		if err != nil {
			return errors.Wrap(err, "error parsing log level")
		}
		log.SetLevel(logLevel)
		lgr := log.WithModule("main")

		lgr.Info("starting ddrp", "git_commit", version.GitCommit, "git_tag", version.GitTag)
		lgr.Info("opening home directory", "path", configuredHomeDir)
		signer, err := cli.GetSigner(configuredHomeDir)
		if err != nil {
			return errors.Wrap(err, "error opening home directory")
		}

		dbPath := config.ExpandDBPath(configuredHomeDir)
		lgr.Info("opening db", "path", dbPath)
		db, err := store.Open(dbPath)
		if err != nil {
			return err
		}

		blobsPath := config.ExpandBlobsPath(configuredHomeDir)
		lgr.Info("opening blob store", "path", blobsPath)
		bs := blob.NewStore(blobsPath)

		seedsStr := cfg.P2P.FixedSeeds
		seeds, err := p2p.ParseSeedPeers(seedsStr)
		if err != nil {
			return errors.Wrap(err, "error parsing seed peers")
		}

		var dnsSeeds []string
		if len(cfg.P2P.DNSSeeds) != 0 {
			seenSeeds := make(map[string]bool)
			for _, domain := range cfg.P2P.DNSSeeds {
				lgr.Info("looking up DNS seeds", "domain", domain)
				seeds, err := p2p.ResolveDNSSeeds(domain)
				if err != nil {
					lgr.Error("error resolving DNS seeds", "domain", domain)
					continue
				}
				for _, seed := range seeds {
					if seenSeeds[seed] == true {
						continue
					}
					seenSeeds[seed] = true
					dnsSeeds = append(dnsSeeds, seed)
				}
			}
		}

		lgr.Info("ingesting ban lists")
		if len(cfg.BanLists) > 0 {
			if err := protocol.IngestBanLists(db, bs, cfg.BanLists); err != nil {
				return errors.Wrap(err, "failed to ingest ban lists")
			}
		}

		var services []service.Service
		mux := p2p.NewPeerMuxer(p2p.MainnetMagic, signer)
		pmCfg := &p2p.PeerManagerOpts{
			Mux:         mux,
			DB:          db,
			SeedPeers:   seeds,
			Signer:      signer,
			ListenHost:  p2pHost,
			MaxInbound:  cfg.P2P.MaxInboundPeers,
			MaxOutbound: cfg.P2P.MaxOutboundPeers,
		}
		pm := p2p.NewPeerManager(pmCfg)
		services = append(services, pm)

		if p2pHost != "" && p2pHost != "127.0.0.1" {
			services = append(services, p2p.NewListener(p2pHost, pm))
		}
		c := client.NewClient(
			cfg.HNSResolver.Host,
			client.WithAPIKey(cfg.HNSResolver.APIKey),
			client.WithPort(cfg.HNSResolver.Port),
			client.WithBasePath(cfg.HNSResolver.BasePath),
		)

		lgr.Info("connecting to HSD", "host", cfg.HNSResolver.Host)
		maxHSDRetries := 10
		for i := 0; i < maxHSDRetries; maxHSDRetries++ {
			if _, err := c.GetInfo(); err != nil {
				lgr.Warn("error connecting to HSD, retrying in 10 seconds", "err", err)
				if i == maxHSDRetries-1 {
					return fmt.Errorf("could not connect to HSD after %d retries", maxHSDRetries)
				}
				time.Sleep(10 * time.Second)
				continue
			}
			break
		}

		nameLocker := util.NewMultiLocker()
		ownPeerID := crypto.HashPub(signer.Pub())

		importer := protocol.NewNameImporter(c, db)
		importer.ConfirmationDepth = cfg.Tuning.NameImporter.ConfirmationDepth
		importer.CheckInterval = config.ConvertDuration(cfg.Tuning.NameImporter.CheckIntervalMS, time.Millisecond)
		importer.Workers = cfg.Tuning.NameImporter.Workers
		importer.VerificationThreshold = cfg.Tuning.NameImporter.VerificationThreshold

		updateQueue := protocol.NewUpdateQueue(mux, db)
		updateQueue.MaxLen = int32(cfg.Tuning.UpdateQueue.MaxLen)
		updateQueue.MinUpdateInterval = config.ConvertDuration(cfg.Tuning.Timebank.MinUpdateIntervalMS, time.Millisecond)

		updater := protocol.NewUpdater(mux, db, updateQueue, nameLocker, bs)
		updater.PollInterval = config.ConvertDuration(cfg.Tuning.Updater.PollIntervalMS, time.Millisecond)
		updater.Workers = cfg.Tuning.Updater.Workers

		pinger := protocol.NewPinger(mux)

		sectorServer := protocol.NewSectorServer(mux, db, bs, nameLocker)
		sectorServer.CacheExpiry = config.ConvertDuration(cfg.Tuning.SectorServer.CacheExpiryMS, time.Millisecond)

		updateServer := protocol.NewUpdateServer(mux, db, nameLocker)

		peerExchanger := protocol.NewPeerExchanger(pm, mux, db)
		peerExchanger.SampleSize = cfg.Tuning.PeerExchanger.SampleSize
		peerExchanger.ResponseTimeout = config.ConvertDuration(cfg.Tuning.PeerExchanger.ResponseTimeoutMS, time.Millisecond)
		peerExchanger.RequestInterval = config.ConvertDuration(cfg.Tuning.PeerExchanger.RequestIntervalMS, time.Millisecond)

		nameSyncer := protocol.NewNameSyncer(mux, db, nameLocker, updater)
		nameSyncer.Workers = cfg.Tuning.NameSyncer.Workers
		nameSyncer.SampleSize = cfg.Tuning.NameSyncer.SampleSize
		nameSyncer.UpdateResponseTimeout = config.ConvertDuration(cfg.Tuning.NameSyncer.UpdateResponseTimeoutMS, time.Millisecond)
		nameSyncer.Interval = config.ConvertDuration(cfg.Tuning.NameSyncer.IntervalMS, time.Millisecond)
		nameSyncer.SyncResponseTimeout = config.ConvertDuration(cfg.Tuning.NameSyncer.SyncResponseTimeoutMS, time.Millisecond)

		server := rpc.NewServer(&rpc.Opts{
			PeerID:      ownPeerID,
			Mux:         mux,
			DB:          db,
			BlobStore:   bs,
			PeerManager: pm,
			NameLocker:  nameLocker,
			Host:        rpcHost,
			Port:        rpcPort,
		})
		services = append(services, []service.Service{
			importer,
			updateQueue,
			updater,
			pinger,
			sectorServer,
			updateServer,
			peerExchanger,
			nameSyncer,
			server,
		}...)

		if cfg.Heartbeat.URL != "" {
			hb := protocol.NewHeartbeater(cfg.Heartbeat.URL, cfg.Heartbeat.Moniker, ownPeerID)
			services = append(services, hb)
		}

		lgr.Info("starting services")
		for _, s := range services {
			go func(s service.Service) {
				if err := s.Start(); err != nil {
					lgr.Error("failed to start service", "err", err)
				}
			}(s)
		}

		if cfg.EnableProfiler {
			lgr.Info("starting profiler", "port", 6060)
			runtime.SetBlockProfileRate(1)
			runtime.SetMutexProfileFraction(1)
			go func() {
				err := http.ListenAndServe("localhost:6060", nil)
				lgr.Error("error starting profiler", "err", err)
			}()
		}

		lgr.Info("dialing seed peers")
		for _, seed := range seeds {
			if err := pm.DialPeer(seed.ID, seed.IP, true); err != nil {
				lgr.Warn("error dialing seed peer", "err", err)
				continue
			}
		}
		for _, seed := range dnsSeeds {
			if err := pm.DialPeer(crypto.ZeroHash, seed, false); err != nil {
				lgr.Warn("error dialing DNS seed peer", "err", err)
			}
		}

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		sig := <-sigs
		lgr.Info("shutting down", "signal", sig)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
