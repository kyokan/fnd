package p2p

import (
	"context"
	"errors"
	"fnd/crypto"
	"fnd/wire"
	"time"
)

func WriteEnvelope(ctx context.Context, peer Peer, signer crypto.Signer, magic uint32, message wire.Message) error {
	envelope, err := wire.NewEnvelope(magic, message, signer)
	if err != nil {
		return err
	}

	return peer.SendCtx(ctx, envelope)
}

var (
	ErrInvalidEnvelopeMagic     = errors.New("envelope has invalid magic")
	ErrInvalidEnvelopeTimestamp = errors.New("envelope has stale timestamp")
	ErrInvalidEnvelopeSignature = errors.New("envelope has invalid signature")
)

func ValidateEnvelope(expMagic uint32, expPeerID crypto.Hash, envelope *wire.Envelope) error {
	now := time.Now()
	lowTime := now.Add(-1 * ValidMessageDeadline)
	highTime := now.Add(ValidMessageDeadline)
	if envelope.Magic != expMagic {
		return ErrInvalidEnvelopeMagic
	}
	if envelope.Timestamp.Before(lowTime) || envelope.Timestamp.After(highTime) {
		return ErrInvalidEnvelopeTimestamp
	}
	if !crypto.VerifySigHashedPub(expPeerID, envelope.Signature, envelope) {
		return ErrInvalidEnvelopeSignature
	}
	return nil
}
