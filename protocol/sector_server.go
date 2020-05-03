package protocol

import (
	"ddrp/blob"
	"ddrp/config"
	"ddrp/crypto"
	"ddrp/log"
	"ddrp/p2p"
	"ddrp/store"
	"ddrp/util"
	"ddrp/wire"
	"errors"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"time"
)

type SectorServer struct {
	CacheExpiry time.Duration
	mux         *p2p.PeerMuxer
	db          *leveldb.DB
	bs          blob.Store
	nameLocker  util.MultiLocker
	lgr         log.Logger
	cache       *util.Cache
}

func NewSectorServer(mux *p2p.PeerMuxer, db *leveldb.DB, bs blob.Store, nameLocker util.MultiLocker) *SectorServer {
	return &SectorServer{
		CacheExpiry: config.ConvertDuration(config.DefaultConfig.Tuning.SectorServer.CacheExpiryMS, time.Millisecond),
		mux:         mux,
		db:          db,
		bs:          bs,
		nameLocker:  nameLocker,
		cache:       util.NewCache(),
		lgr:         log.WithModule("sector-server"),
	}
}

func (s *SectorServer) Start() error {
	s.mux.AddMessageHandler(p2p.PeerMessageHandlerForType(wire.MessageTypeTreeBaseReq, s.onTreeBaseReq))
	s.mux.AddMessageHandler(p2p.PeerMessageHandlerForType(wire.MessageTypeSectorReq, s.onSectorReq))
	return nil
}

func (s *SectorServer) Stop() error {
	return nil
}

func (s *SectorServer) onTreeBaseReq(peerID crypto.Hash, envelope *wire.Envelope) {
	reqMsg := envelope.Message.(*wire.TreeBaseReq)
	lgr := s.lgr.Sub(
		"name", reqMsg.Name,
		"peer_id", peerID,
	)

	if !s.nameLocker.TryRLock(reqMsg.Name) {
		lgr.Info("dropping diff req for busy name")
		return
	}
	merkleBase, err := store.GetMerkleBase(s.db, reqMsg.Name)
	if err != nil {
		s.nameLocker.RUnlock(reqMsg.Name)
		lgr.Error("error getting merkle base", "err", err)
		return
	}
	s.nameLocker.RUnlock(reqMsg.Name)

	resMsg := &wire.TreeBaseRes{
		Name:       reqMsg.Name,
		MerkleBase: merkleBase,
	}
	if err := s.mux.Send(peerID, resMsg); err != nil {
		lgr.Error("error serving tree base response", "err", err)
		return
	}
	lgr.Debug("served tree base response")
}

func (s *SectorServer) onSectorReq(peerID crypto.Hash, envelope *wire.Envelope) {
	reqMsg := envelope.Message.(*wire.SectorReq)
	lgr := s.lgr.Sub(
		"name", reqMsg.Name,
		"peer_id", peerID,
	)

	if !s.nameLocker.TryRLock(reqMsg.Name) {
		lgr.Info("dropping sector req for busy name")
		return
	}
	header, err := store.GetHeader(s.db, reqMsg.Name)
	if errors.Is(err, leveldb.ErrNotFound) {
		s.nameLocker.RUnlock(reqMsg.Name)
		return
	}
	if err != nil {
		lgr.Error("error getting blob header", "err", err)
		s.nameLocker.RUnlock(reqMsg.Name)
		return
	}
	cacheKey := fmt.Sprintf("%s:%d:%d", reqMsg.Name, header.Timestamp.Unix(), reqMsg.SectorID)
	cached := s.cache.Get(cacheKey)
	if cached != nil {
		s.nameLocker.RUnlock(reqMsg.Name)
		s.sendResponse(peerID, reqMsg.Name, reqMsg.SectorID, cached.(blob.Sector))
		return
	}

	bl, err := s.bs.Open(reqMsg.Name)
	if err != nil {
		s.nameLocker.RUnlock(reqMsg.Name)
		lgr.Error(
			"failed to fetch blob",
			"err", err,
		)
		return
	}
	defer func() {
		if err := bl.Close(); err != nil {
			s.lgr.Error("failed to close blob", "err", err)
		}
	}()
	sector, err := bl.ReadSector(reqMsg.SectorID)
	if err != nil {
		s.nameLocker.RUnlock(reqMsg.Name)
		lgr.Error(
			"failed to read sector",
			"err", err,
		)
		return
	}
	s.cache.Set(cacheKey, sector, int64(s.CacheExpiry/time.Millisecond))
	s.nameLocker.RUnlock(reqMsg.Name)
	s.sendResponse(peerID, reqMsg.Name, reqMsg.SectorID, sector)
}

func (s *SectorServer) sendResponse(peerID crypto.Hash, name string, sectorID uint8, sector blob.Sector) {
	resMsg := &wire.SectorRes{
		Name:     name,
		SectorID: sectorID,
		Sector:   sector,
	}
	if err := s.mux.Send(peerID, resMsg); err != nil {
		s.lgr.Error("error serving sector response", "err", err)
		return
	}
	s.lgr.Debug(
		"served sector response",
		"peer_id", peerID,
		"sector_id", sectorID,
	)
}
