package p2p

import (
	"context"
	"ddrp/crypto"
	"ddrp/version"
	"ddrp/wire"
	"github.com/pkg/errors"
	"time"
)

var (
	ErrUnexpectedMessage    = errors.New("unexpected handshake message")
	ErrIncompatibleProtocol = errors.New("incompatible protocol version")
	ErrInvalidNonce         = errors.New("invalid nonce on hello message")
)

type HandshakeConfig struct {
	Magic           uint32
	ProtocolVersion uint32
	Peer            Peer
	Signer          crypto.Signer
}

func HandleOutgoingHandshake(ctx context.Context, cfg *HandshakeConfig) (crypto.Hash, error) {
	localNonce := crypto.Rand32()
	ourHelloMsg := &wire.Hello{
		ProtocolVersion: cfg.ProtocolVersion,
		LocalNonce:      localNonce,
		PublicKey:       cfg.Signer.Pub(),
		UserAgent:       version.UserAgent,
	}

	err := WriteEnvelope(ctx, cfg.Peer, cfg.Signer, cfg.Magic, ourHelloMsg)
	if err != nil {
		return crypto.ZeroHash, errors.Wrap(err, "failed to send hello message")
	}

	subCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	theirHelloEnv, err := cfg.Peer.ReceiveCtx(subCtx)
	if err != nil {
		return crypto.ZeroHash, errors.Wrap(err, "failed to receive peer hello message")
	}
	if theirHelloEnv.MessageType != wire.MessageTypeHello {
		return crypto.ZeroHash, ErrUnexpectedMessage
	}
	theirHelloMsg := theirHelloEnv.Message.(*wire.Hello)
	theirPeerID := crypto.HashPub(theirHelloMsg.PublicKey)
	if err := ValidateEnvelope(cfg.Magic, theirPeerID, theirHelloEnv); err != nil {
		return crypto.ZeroHash, errors.Wrap(err, "peer responded with invalid hello message")
	}
	if theirHelloMsg.ProtocolVersion > cfg.ProtocolVersion {
		return crypto.ZeroHash, ErrIncompatibleProtocol
	}
	if theirHelloMsg.RemoteNonce != localNonce {
		return crypto.ZeroHash, ErrInvalidNonce
	}

	remoteNonce := theirHelloMsg.LocalNonce
	ourHelloAckMsg := &wire.HelloAck{
		Nonce: remoteNonce,
	}
	if err := WriteEnvelope(ctx, cfg.Peer, cfg.Signer, cfg.Magic, ourHelloAckMsg); err != nil {
		return crypto.ZeroHash, errors.Wrap(err, "failed to send hello ack message")
	}
	return theirPeerID, nil
}
