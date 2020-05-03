package protocol

import (
	"ddrp/blob"
	"ddrp/store"
	"ddrp/testutil/testfs"
	"github.com/stretchr/testify/require"
	"testing"
)

// TODO: refactor to use the httptest server

func TestIngestBanLists(t *testing.T) {
	db, doneDB := setupDB(t)
	defer doneDB()
	tmpDir, doneDir := testfs.NewTempDir(t)
	defer doneDir()

	bs := blob.NewStore(tmpDir)
	bl, err := bs.Open("foo")
	require.NoError(t, err)
	require.NoError(t, bl.Close())

	err = IngestBanLists(db, bs, []string{
		"https://gist.githubusercontent.com/mslipper/b5d28fe54850be8b879b4064abcacddf/raw/6c412b8bc02a3765b49acf455b6ce32c7a5a4ddd/banlist-test-1",
		"https://gist.githubusercontent.com/mslipper/cc32c19da426622e156cf771aafa58da/raw/07210f958c01737a139d916c5082dd1bc7724955/banlist-test-2",
	})
	require.NoError(t, err)

	fooBanned, err := store.NameIsBanned(db, "foo")
	require.NoError(t, err)
	require.True(t, fooBanned)
	barBanned, err := store.NameIsBanned(db, "bar")
	require.NoError(t, err)
	require.True(t, barBanned)
	bazBanned, err := store.NameIsBanned(db, "baz")
	require.NoError(t, err)
	require.False(t, bazBanned)

	exists, err := bs.Exists("foo")
	require.NoError(t, err)
	require.False(t, exists)
}
