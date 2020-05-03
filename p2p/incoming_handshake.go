package p2p

import (
	"context"
	"ddrp/crypto"
	"ddrp/wire"
	"github.com/pkg/errors"
)

func HandleIncomingHandshake(ctx context.Context, cfg *HandshakeConfig) (crypto.Hash, error) {
	localNonce := crypto.Rand32()
	theirHelloEnv, err := cfg.Peer.ReceiveCtx(ctx)
	if err != nil {
		return crypto.ZeroHash, errors.Wrap(err, "failed to receive hello message")
	}
	if theirHelloEnv.MessageType != wire.MessageTypeHello {
		return crypto.ZeroHash, ErrUnexpectedMessage
	}
	theirHelloMsg := theirHelloEnv.Message.(*wire.Hello)
	theirPeerID := crypto.HashPub(theirHelloMsg.PublicKey)
	if err := ValidateEnvelope(cfg.Magic, crypto.HashPub(theirHelloMsg.PublicKey), theirHelloEnv); err != nil {
		return crypto.ZeroHash, errors.Wrap(err, "peer initiated with invalid hello message")
	}
	if theirHelloMsg.ProtocolVersion > cfg.ProtocolVersion {
		return crypto.ZeroHash, ErrIncompatibleProtocol
	}

	ourHelloMsg := &wire.Hello{
		ProtocolVersion: cfg.ProtocolVersion,
		LocalNonce:      localNonce,
		RemoteNonce:     theirHelloMsg.LocalNonce,
		PublicKey:       cfg.Signer.Pub(),
	}
	if err := WriteEnvelope(ctx, cfg.Peer, cfg.Signer, cfg.Magic, ourHelloMsg); err != nil {
		return crypto.ZeroHash, errors.Wrap(err, "failed to respond with hello message")
	}

	theirHelloAckEnv, err := cfg.Peer.ReceiveCtx(ctx)
	if err != nil {
		return crypto.ZeroHash, errors.Wrap(err, "failed to receive hello ack message")
	}
	if theirHelloAckEnv.MessageType != wire.MessageTypeHelloAck {
		return crypto.ZeroHash, ErrUnexpectedMessage
	}
	if err := ValidateEnvelope(cfg.Magic, theirPeerID, theirHelloAckEnv); err != nil {
		return crypto.ZeroHash, errors.Wrap(err, "peer responded with invalid hello ack message")
	}
	theirHelloAck := theirHelloAckEnv.Message.(*wire.HelloAck)
	if theirHelloAck.Nonce != localNonce {
		return crypto.ZeroHash, ErrInvalidNonce
	}

	return theirPeerID, nil
}
