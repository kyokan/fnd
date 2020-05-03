package protocol

import (
	"bytes"
	"ddrp/crypto"
	"ddrp/log"
	"ddrp/service"
	"ddrp/version"
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
	"sync"
	"time"
)

const (
	DefaultInterval = 30 * time.Second
	DefaultTimeout  = 10 * time.Second
)

type Heartbeat struct {
	Moniker   string      `json:"moniker"`
	PeerID    crypto.Hash `json:"peer_id"`
	UserAgent string      `json:"user_agent"`
}

type Heartbeater struct {
	Interval time.Duration
	Timeout  time.Duration
	url      string
	moniker  string
	peerID   crypto.Hash
	lgr      log.Logger
	quitCh   chan struct{}
	once     sync.Once
}

var _ service.Service = (*Heartbeater)(nil)

func NewHeartbeater(url string, moniker string, peerID crypto.Hash) *Heartbeater {
	srv := &Heartbeater{
		Interval: DefaultInterval,
		Timeout:  DefaultTimeout,
		url:      url,
		moniker:  moniker,
		peerID:   peerID,
		lgr:      log.WithModule("heartbeat"),
		quitCh:   make(chan struct{}),
	}

	return srv
}

func (s *Heartbeater) Start() error {
	client := &http.Client{
		Timeout: s.Timeout,
	}

	beat := &Heartbeat{
		Moniker:   s.moniker,
		PeerID:    s.peerID,
		UserAgent: version.UserAgent,
	}
	beatJSON, err := json.Marshal(beat)
	if err != nil {
		return errors.Wrap(err, "failed to marshal heartbeat")
	}

	ticker := time.NewTicker(s.Interval)
	for {
		select {
		case <-ticker.C:
			res, err := client.Post(s.url, "application/json", bytes.NewReader(beatJSON))
			if err != nil {
				s.lgr.Error("failed to send heartbeat", "err", err)
				continue
			}
			if err := res.Body.Close(); err != nil {
				s.lgr.Error("failed to close response body", "err", err)
			}
			if res.StatusCode != 204 {
				s.lgr.Warn("heartbeat server sent non-204 status code", "code", res.StatusCode)
			}
		case <-s.quitCh:
			return nil
		}
	}
}

func (s *Heartbeater) Stop() error {
	close(s.quitCh)
	return nil
}
