package protocol

import (
	"bytes"
	"encoding/hex"
	"encoding/base64"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"fnd/config"
	"fnd/log"
	"fnd/store"
	"fnd.localhost/handshake/client"
	"fnd.localhost/handshake/dns"
	"fnd.localhost/handshake/primitives"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
	"sync/atomic"
	"time"
)

type NameImporter struct {
	ConfirmationDepth     int
	CheckInterval         time.Duration
	Workers               int
	VerificationThreshold float64

	client *client.Client
	db     *leveldb.DB
	lgr    log.Logger
	quitCh chan struct{}
}

type HNSName struct {
	Name      string
	PublicKey *btcec.PublicKey
}

func NewNameImporter(client *client.Client, db *leveldb.DB) *NameImporter {
	return &NameImporter{
		ConfirmationDepth:     config.DefaultConfig.Tuning.NameImporter.ConfirmationDepth,
		CheckInterval:         config.ConvertDuration(config.DefaultConfig.Tuning.NameImporter.CheckIntervalMS, time.Millisecond),
		Workers:               config.DefaultConfig.Tuning.NameImporter.Workers,
		VerificationThreshold: config.DefaultConfig.Tuning.NameImporter.VerificationThreshold,
		client:                client,
		db:                    db,
		lgr:                   log.WithModule("hns-importer"),
		quitCh:                make(chan struct{}, 1),
	}
}

func (n *NameImporter) Start() error {
	return n.sync()
}

func (n *NameImporter) Stop() error {
	close(n.quitCh)
	return nil
}

func (n *NameImporter) sync() error {
	ticker := time.NewTicker(n.CheckInterval)

	for {
		n.doSync()

		select {
		case <-ticker.C:
		case <-n.quitCh:
			return nil
		}
	}
}

func (n *NameImporter) doSync() {
	info, err := n.client.RPCGetBlockchainInfo()
	if err != nil {
		n.lgr.Error("failed to get chain info", "err", err)
		return
	}
	if info.VerificationProgress < n.VerificationThreshold {
		n.lgr.Info("chain not synced, trying again later", "progress", info.VerificationProgress)
		return
	}

	chainHeight := info.Blocks
	if chainHeight < 0 {
		n.lgr.Error("chain height is negative")
		return
	}
	if chainHeight < n.ConfirmationDepth {
		n.lgr.Info("chain height less than confirmation count, skipping name import")
		return
	}
	syncedHeight, err := store.GetLastNameImportHeight(n.db)
	if err != nil {
		n.lgr.Error("failed to get synced height", "err", err)
		return
	}
	confirmedHeight := chainHeight - n.ConfirmationDepth
	if confirmedHeight == syncedHeight {
		n.lgr.Info("fully synced, skipping name import", "synced_height", syncedHeight)
		return
	}
	if confirmedHeight < syncedHeight {
		n.lgr.Warn("confirmed chain height behind synced height", "confirmed_height", confirmedHeight, "synced_height", syncedHeight)
		return
	}

	n.lgr.Info("importing blocks", "start_height", syncedHeight+1, "end_height", confirmedHeight)
	var importCount int
	for height := syncedHeight + 1; height < confirmedHeight; {
		delta := confirmedHeight - height
		if delta > n.Workers {
			delta = n.Workers
		}

		blocks, err := n.fetchBlocks(height, delta)
		if err != nil {
			n.lgr.Error("error fetching blocks", "err", err)
			return
		}

		for i, block := range blocks {
			blockHeight := height + i
			records := ExtractTXTRecordsBlock(block)
			var names []string
			for _, update := range records {
				name, err := n.client.RPCGetNameByHash(update.NameHash)
				if err != nil || name == nil {
					n.lgr.Error("error resolving name hash", "err", err)
					return
				}
				names = append(names, *name)
			}

			err := store.WithTx(n.db, func(tx *leveldb.Transaction) error {
				for i, record := range records {
					if err := store.SetNameInfoTx(tx, names[i], record.PublicKey, blockHeight); err != nil {
						return errors.Wrap(err, "error inserting name info")
					}
				}
				if err := store.SetLastNameImportHeightTx(tx, blockHeight); err != nil {
					return errors.Wrap(err, "error setting last name import height")
				}
				if height+i == confirmedHeight-1 {
					if err := store.SetInitialImportCompleteTx(tx); err != nil {
						return errors.Wrap(err, "error setting initial import complete")
					}
				}
				return nil
			})
			if err != nil {
				n.lgr.Error("error processing block", "height", blockHeight, "err", err)
				return
			}
			n.lgr.Info("processed block", "height", blockHeight)
		}
		height += len(blocks)
	}

	n.lgr.Info("import complete", "import_count", importCount)
}

func (n *NameImporter) fetchBlocks(start int, count int) ([]*primitives.Block, error) {
	partition := make([]*primitives.Block, count)
	var wg sync.WaitGroup
	var workerErr atomic.Value
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			blockHex, err := n.client.RPCGetBlockHexByHeight(start + i)
			if err != nil {
				workerErr.Store(errors.Wrap(err, fmt.Sprintf("error fetching HNS block %d", start+i)))
				return
			}
			blockB, err := hex.DecodeString(blockHex)
			if err != nil {
				workerErr.Store(errors.Wrap(err, fmt.Sprintf("error parsing HNS block hex %d", start+i)))
				return
			}
			block := new(primitives.Block)
			if err := block.Decode(bytes.NewReader(blockB)); err != nil {
				workerErr.Store(errors.Wrap(err, fmt.Sprintf("error decoding HNS block %d", start+i)))
				return
			}
			partition[i] = block
		}(i)
	}
	wg.Wait()
	err := workerErr.Load()
	if err != nil {
		return nil, errors.Wrap(err.(error), "error fetching blocks")
	}
	return partition, nil
}

type FNRecord struct {
	NameHash  string
	PublicKey *btcec.PublicKey
}

func ExtractTXTRecordsBlock(block *primitives.Block) []*FNRecord {
	uniqRecords := make(map[string]*FNRecord)
	var order []string
	for _, tx := range block.Transactions {
		records := ExtractTXTRecordsTx(tx)
		for _, rec := range records {
			if _, ok := uniqRecords[rec.NameHash]; !ok {
				order = append(order, rec.NameHash)
			}
			uniqRecords[rec.NameHash] = rec
		}
	}
	out := make([]*FNRecord, len(uniqRecords))
	for i := 0; i < len(order); i++ {
		out[i] = uniqRecords[order[i]]
	}
	return out
}

func ExtractTXTRecordsTx(tx *primitives.Transaction) []*FNRecord {
	var out []*FNRecord
	for _, vOut := range tx.Outputs {
		covenant := vOut.Covenant
		var resource *dns.Resource
		var nh []byte
		switch covenant.Type {
		case primitives.CovenantUpdate:
			cov, err := primitives.UpdateFromCovenant(covenant)
			if err != nil {
				continue
			}
			resource = cov.Resource
			nh = cov.NameHash
		case primitives.CovenantRegister:
			cov, err := primitives.RegisterFromCovenant(covenant)
			if err != nil {
				continue
			}
			resource = cov.Resource
			nh = cov.NameHash
		default:
			continue
		}

		var pub *btcec.PublicKey
		if resource == nil {
			continue
		}
		for _, record := range resource.Records {
			txt, ok := record.(*dns.TXTRecord)
			if !ok {
				continue
			}
			for _, entry := range txt.Entries {
				p, err := ParseFNRecord(entry)
				if err != nil {
					continue
				}
				pub = p
			}
		}

		if pub == nil {
			continue
		}

		out = append(out, &FNRecord{
			NameHash:  hex.EncodeToString(nh),
			PublicKey: pub,
		})
	}
	return out
}

func ParseFNRecord(record string) (*btcec.PublicKey, error) {
	if len(record) != 45 {
		return nil, errors.New("mal-formed txt record")
	}
	prefix := record[0:1]
	pubkeyb64 := record[1:]
	if prefix != "f" {
		return nil, errors.New("mal-formed record sigil")
	}
	if len(pubkeyb64) != 44 {
		return nil, errors.New("invalid public key length")
	}
	keyBytes, err := base64.StdEncoding.DecodeString(pubkeyb64)
	if err != nil {
		return nil, errors.New("mal-formed public key")
	}
	pub, err := btcec.ParsePubKey(keyBytes, btcec.S256())
	if err != nil {
		return nil, errors.Wrap(err, "error parsing public key")
	}
	return pub, nil
}
