package p2p

import (
	"fmt"
	"fnd/log"
	"fnd/service"
	"net"
	"sync"

	"github.com/pkg/errors"
)

type Listener struct {
	host    string
	port    int
	manager PeerManager
	lgr     log.Logger
	quitCh  chan struct{}
	once    sync.Once
}

var _ service.Service = (*Listener)(nil)

func NewListener(host string, manager PeerManager) *Listener {
	return &Listener{
		host:    host,
		manager: manager,
		lgr:     log.WithModule("listener"),
		quitCh:  make(chan struct{}),
	}
}

func (l *Listener) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", l.host, StandardPort))
	if err != nil {
		return err
	}

	go func() {
		<-l.quitCh
		err := listener.Close()
		if err != nil {
			l.lgr.Error("failed to shut down listener", "err", err)
		} else {
			l.lgr.Info("listener shut down")
		}
		return
	}()

	l.lgr.Info("listening for connections", "host", l.host, "port", l.port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return errors.Wrap(err, "error accepting peer connection")
		}
		l.lgr.Info(
			"accepted new peer connection",
			"remote_addr", conn.RemoteAddr(),
		)
		go func() {
			if err := l.manager.AcceptPeer(conn.(*net.TCPConn)); err != nil {
				l.lgr.Info(
					"peer connection rejected",
					"remote_addr", conn.RemoteAddr(),
					"reason", err,
				)
			}
		}()
	}
}

func (l *Listener) Stop() error {
	close(l.quitCh)
	return nil
}
