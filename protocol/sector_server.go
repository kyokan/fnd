package protocol

import (
	"bytes"
	"fmt"
	"fnd/blob"
	"fnd/config"
	"fnd/crypto"
	"fnd/log"
	"fnd/p2p"
	"fnd/store"
	"fnd/util"
	"fnd/wire"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
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
	s.mux.AddMessageHandler(p2p.PeerMessageHandlerForType(wire.MessageTypeBlobReq, s.onBlobReq))
	s.mux.AddMessageHandler(p2p.PeerMessageHandlerForType(wire.MessageTypeEquivocationProof, s.onEquivocationProof))
	return nil
}

func (s *SectorServer) Stop() error {
	return nil
}

// onBlobReq handles a BlobReq and responds with a BlobRes
//
// In case the BlobReq contains SectorSize == MaxSectors, it is considered a
// special case, which is a part of the equivocation proof flow. Note that in a
// normal BlobReq/BlobRes flow, SectorSize may never be MaxSectors as it is the
// start for the expected payload.
//
// In the special case where SectorSize = MaxSectors, the BlobReq is considered
// an Equivocation Request. In response, the peer is sent an Equivocation Proof.
//
// Equivocation Proof flow:
// Syncer - sees a conflicting update
// - Writes equivocation proof locally to db
// - Sends an equivocation Update (special case where Update.SectorSize = 0)
// - Expects peers to reply with Equivocation Request i.e. BlobReq where SectorSize = MaxSectors
// - Responds with Equivocation Proof
func (s *SectorServer) onBlobReq(peerID crypto.Hash, envelope *wire.Envelope) {
	reqMsg := envelope.Message.(*wire.BlobReq)
	lgr := s.lgr.Sub(
		"name", reqMsg.Name,
		"peer_id", peerID,
	)

	header, err := store.GetHeader(s.db, reqMsg.Name)
	if err != nil {
		lgr.Error(
			"failed to fetch header",
			"err", err)
		return
	}

	if header.EpochHeight < reqMsg.EpochHeight {
		lgr.Error(
			"request epoch height must be <= header epoch height",
			"request_epoch_height", reqMsg.EpochHeight,
			"header_epoch_height", header.EpochHeight)
		return
	}

	// Blob request with SectorSize == blob.MaxSectors is an equivocation
	// request. Respond with equivocation proof.
	if reqMsg.SectorSize == blob.MaxSectors {
		raw, err := store.GetEquivocationProof(s.db, reqMsg.Name)
		if err != nil {
			lgr.Error(
				"failed to fetch equivocation proof",
				"err", err)
			return
		}
		proof := &wire.EquivocationProof{}
		buf := bytes.NewReader(raw)
		if err := proof.Decode(buf); err != nil {
			lgr.Error(
				"failed to deserialize equivocation proof",
				"err", err)
			return
		}
		if err := s.mux.Send(peerID, proof); err != nil {
			s.lgr.Error("error serving equivocation proof", "err", err)
			return
		}
		return
	}

	if !s.nameLocker.TryRLock(reqMsg.Name) {
		lgr.Info("dropping sector req for busy name")
		return
	}

	var prevHash crypto.Hash = blob.ZeroHash
	if reqMsg.SectorSize != 0 {
		hash, err := store.GetSectorHash(s.db, reqMsg.Name, reqMsg.SectorSize-1)
		if err != nil {
			lgr.Error(
				"failed to fetch sector hash",
				"err", err)
			return
		}
		prevHash = hash
	}

	cacheKey := fmt.Sprintf("%s:%d:%d", reqMsg.Name, reqMsg.EpochHeight, reqMsg.SectorSize)
	cached := s.cache.Get(cacheKey)
	if cached != nil {
		s.nameLocker.RUnlock(reqMsg.Name)
		s.sendResponse(peerID, reqMsg.Name, prevHash, cached.([]blob.Sector), header.EpochHeight, reqMsg.SectorSize, header.Signature)
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
	var sectors []blob.Sector
	for i := reqMsg.SectorSize; i < header.SectorSize; i++ {
		sector := &blob.Sector{}
		_, err = bl.ReadAt(sector[:], int64(i)*blob.SectorBytes)
		if err != nil {
			s.nameLocker.RUnlock(reqMsg.Name)
			lgr.Error(
				"failed to read sector",
				"err", err,
			)
			return
		}
		sectors = append(sectors, *sector)
	}
	s.cache.Set(cacheKey, sectors, int64(s.CacheExpiry/time.Millisecond))
	s.nameLocker.RUnlock(reqMsg.Name)
	s.sendResponse(peerID, reqMsg.Name, prevHash, sectors, header.EpochHeight, reqMsg.SectorSize, header.Signature)
}

func (s *SectorServer) sendResponse(peerID crypto.Hash, name string, prevHash crypto.Hash, sectors []blob.Sector, epochHeight, sectorSize uint16, signature crypto.Signature) {
	resMsg := &wire.BlobRes{
		Name:            name,
		EpochHeight:     epochHeight,
		PayloadPosition: sectorSize,
		PrevHash:        prevHash,
		Payload:         sectors,
		Signature:       signature,
	}
	if err := s.mux.Send(peerID, resMsg); err != nil {
		s.lgr.Error("error serving sector response", "err", err)
		return
	}
	s.lgr.Debug(
		"served sector response",
		"peer_id", peerID,
		"sector_size", sectorSize,
	)
}

func (s *SectorServer) onEquivocationProof(peerID crypto.Hash, envelope *wire.Envelope) {
	msg := envelope.Message.(*wire.EquivocationProof)
	lgr := s.lgr.Sub(
		"name", msg.Name,
		"peer_id", peerID,
	)

	lgr.Trace("handling equivocation response", "name", msg.Name)
	if msg.LocalEpochHeight != msg.RemoteEpochHeight {
		s.lgr.Warn("unexpected epoch height", "local_epoch_height", msg.LocalEpochHeight, "remote_epoch_height", msg.RemoteEpochHeight)
		return
	}

	header, err := store.GetHeader(s.db, msg.Name)
	if err != nil {
		lgr.Error(
			"failed to fetch header",
			"err", err)
		return
	}

	bannedAt, err := store.GetHeaderBan(s.db, msg.Name)
	if err != nil {
		lgr.Error(
			"failed to fetch header ban",
			"err", err)
		return
	}

	// Skip if name is already in equivocated state.
	if !bannedAt.IsZero() && bannedAt.Add(7*24*time.Duration(time.Hour)).After(time.Now()) {
		s.lgr.Warn("name banned", "name", msg.Name)
		return
	}

	// Skip if equivocation proof was already used to ban.
	if !bannedAt.IsZero() && header.EpochHeight == msg.RemoteEpochHeight {
		s.lgr.Warn("already equivocated", "name", msg.Name)
		return
	}

	if header.EpochHeight > msg.RemoteEpochHeight {
		s.lgr.Warn("remote epoch height must be <= header epoch heigth", "header_epoch_height", header.EpochHeight, "remote_epoch_height", msg.RemoteEpochHeight)
		return
	}

	if msg.LocalSectorSize != msg.RemotePayloadPosition {
		s.lgr.Warn("unexpected sector size", "local_sector_size", msg.LocalSectorSize, "remote_payload_position", msg.RemotePayloadPosition)
		return
	}
	if err := validateBlobUpdate(s.db, msg.Name, msg.LocalEpochHeight, msg.LocalSectorSize, msg.LocalSectorTipHash, msg.LocalReservedRoot, msg.LocalSignature); err != nil {
		lgr.Warn("local signaure validation failed", "err", err)
		return
	}
	sectorSize := msg.RemotePayloadPosition + uint16(len(msg.RemotePayload))
	// Additional sanity check: make sure that update does not overflow max sectors.
	if int(sectorSize) > blob.MaxSectors {
		lgr.Warn("received unexpected sector size", "sector_size", sectorSize, "max", blob.MaxSectors)
		return
	}
	// Generate the current tip hash from prev hash and the payload
	// sectors.
	var sectorTipHash crypto.Hash = msg.RemotePrevHash
	for i := 0; int(i) < len(msg.RemotePayload); i++ {
		sectorTipHash = blob.SerialHashSector(msg.RemotePayload[i], sectorTipHash)
	}
	// Verify that the update is valid by using the recomputed
	// sector size, sector tip hash and other metadata. This data
	// is first hashed and the signature is validated against the
	// name's pubkey. See validateBlobRes.
	if err := validateBlobUpdate(s.db, msg.Name, msg.RemoteEpochHeight, sectorSize, sectorTipHash, msg.RemoteReservedRoot, msg.RemoteSignature); err != nil {
		lgr.Warn("remote signaure validation failed", "err", err)
		return
	}
	lgr.Trace("equivocation proof valid ", "name", msg.Name)
	if err := store.WithTx(s.db, func(tx *leveldb.Transaction) error {
		return store.SetEquivocationProofTx(tx, msg.Name, msg)
	}); err != nil {
		lgr.Trace("error writing equivocation proof", "err", err)
	}
	err = store.WithTx(s.db, func(tx *leveldb.Transaction) error {
		return store.SetHeaderBan(tx, msg.Name, time.Time{})
	})
	if err != nil {
		lgr.Warn("error setting header banned", "err", err)
		return
	}
	return
}
