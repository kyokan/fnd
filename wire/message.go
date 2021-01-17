package wire

import (
	"io"

	"fnd/crypto"
	"fnd/dwire"
)

type Message interface {
	crypto.Hasher
	dwire.EncodeDecoder
	MsgType() MessageType
	Equals(other Message) bool
}

type MessageType uint16

const (
	MessageTypeHello MessageType = iota
	MessageTypeHelloAck
	MessageTypePing
	MessageTypeUpdate
	MessageTypeNilUpdate
	MessageTypeBlobReq
	MessageTypeBlobRes
	MessageTypePeerReq
	MessageTypePeerRes
	MessageTypeUpdateReq
	MessageTypeNameRes
)

func (t MessageType) String() string {
	switch t {
	case MessageTypeHello:
		return "Hello"
	case MessageTypeHelloAck:
		return "HelloAck"
	case MessageTypePing:
		return "Ping"
	case MessageTypeUpdate:
		return "Update"
	case MessageTypeNilUpdate:
		return "NilUpdate"
	case MessageTypeBlobReq:
		return "BlobReq"
	case MessageTypeBlobRes:
		return "BlobRes"
	case MessageTypePeerReq:
		return "PeerReq"
	case MessageTypePeerRes:
		return "PeerRes"
	case MessageTypeUpdateReq:
		return "UpdateReq"
	case MessageTypeNameRes:
		return "NameRes"
	default:
		return "unknown"
	}
}

func (t MessageType) Encode(w io.Writer) error {
	return dwire.EncodeField(w, uint16(t))
}

func (t *MessageType) Decode(r io.Reader) error {
	var decoded uint16
	if err := dwire.DecodeField(r, &decoded); err != nil {
		return err
	}
	*t = MessageType(decoded)
	return nil
}
