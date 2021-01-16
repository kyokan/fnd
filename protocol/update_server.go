package protocol

import (
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/ddrp-org/ddrp/log"
	"github.com/ddrp-org/ddrp/p2p"
	"github.com/ddrp-org/ddrp/store"
	"github.com/ddrp-org/ddrp/util"
	"github.com/ddrp-org/ddrp/wire"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

type UpdateServer struct {
	mux        *p2p.PeerMuxer
	nameLocker util.MultiLocker
	db         *leveldb.DB
	lgr        log.Logger
}

func NewUpdateServer(mux *p2p.PeerMuxer, db *leveldb.DB, nameLocker util.MultiLocker) *UpdateServer {
	return &UpdateServer{
		mux:        mux,
		db:         db,
		nameLocker: nameLocker,
		lgr:        log.WithModule("update-server"),
	}
}

func (u *UpdateServer) Start() error {
	u.mux.AddMessageHandler(p2p.PeerMessageHandlerForType(wire.MessageTypeUpdateReq, u.UpdateReqHandler))
	return nil
}

func (u *UpdateServer) Stop() error {
	return nil
}

func (u *UpdateServer) UpdateReqHandler(peerID crypto.Hash, envelope *wire.Envelope) {
	msg := envelope.Message.(*wire.UpdateReq)
	u.lgr.Debug("receive update req", "name", msg.Name, "epoch", msg.EpochHeight, "sector", msg.SectorSize)

	if !u.nameLocker.TryRLock(msg.Name) {
		if err := u.mux.Send(peerID, wire.NewNilUpdate(msg.Name)); err != nil {
			u.lgr.Error("error sending response to update req", "name", msg.Name, "err", err)
		} else {
			u.lgr.Debug("serving nil update response for busy name", "name", msg.Name)
		}
		return
	}
	defer u.nameLocker.RUnlock(msg.Name)

	header, err := store.GetHeader(u.db, msg.Name)
	if errors.Is(err, leveldb.ErrNotFound) {
		if err := u.mux.Send(peerID, wire.NewNilUpdate(msg.Name)); err != nil {
			u.lgr.Error("error sending response to update req", "name", msg.Name, "err", err)
		} else {
			u.lgr.Debug("serving nil update response for unknown name", "name", msg.Name)
		}
		return
	}
	if err != nil {
		u.lgr.Error("error reading blob header", "name", msg.Name, "err", err)
		if err := u.mux.Send(peerID, wire.NewNilUpdate(msg.Name)); err != nil {
			u.lgr.Error("error sending response to update req", "name", msg.Name, "err", err)
		} else {
			u.lgr.Debug("serving nil update response for name after error reading header", "name", msg.Name)
		}
		return
	}

	if header.SectorSize < msg.SectorSize || header.SectorSize == msg.SectorSize {
		if err := u.mux.Send(peerID, wire.NewNilUpdate(msg.Name)); err != nil {
			u.lgr.Error("error sending response to update req", "name", msg.Name, "err", err)
		} else {
			u.lgr.Debug("serving nil update response for future header", "name", msg.Name)
		}
		return
	}

	err = u.mux.Send(peerID, &wire.Update{
		Name:          msg.Name,
		EpochHeight:   header.EpochHeight,
		SectorSize:    header.SectorSize,
		SectorTipHash: header.SectorTipHash,
		Signature:     header.Signature,
	})
	if err != nil {
		u.lgr.Error("error serving update", "name", msg.Name, "err", err)
		return
	}

	u.lgr.Debug("served update", "name", msg.Name)
}
