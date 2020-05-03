package protocol

import (
	"ddrp/crypto"
	"ddrp/testutil/testcrypto"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestHeartbeater_SendHeartbeats(t *testing.T) {
	var once sync.Once
	resCh := make(chan *Heartbeat)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() {
			defer r.Body.Close()
			body, err := ioutil.ReadAll(r.Body)
			require.NoError(t, err)
			var beat Heartbeat
			require.NoError(t, json.Unmarshal(body, &beat))
			resCh <- &beat
		})
		w.WriteHeader(204)
	}))

	_, pub := testcrypto.RandKey()
	peerID := crypto.HashPub(pub)
	heartbeater := NewHeartbeater(server.URL, "test moniker", peerID)
	heartbeater.Interval = 250 * time.Millisecond
	go heartbeater.Start()

	beat := <-resCh
	require.Equal(t, "test moniker", beat.Moniker)
	require.Equal(t, peerID, beat.PeerID)
	require.Equal(t, "ddrpd/+", beat.UserAgent)

	require.NoError(t, heartbeater.Stop())
	server.Close()
}
