package postgrestorage

import (
	"github.com/KyberNetwork/reserve-stats/lib/timeutil"
	"github.com/KyberNetwork/reserve-stats/tradelogs/common"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTradeLogDB_GetUserList(t *testing.T) {
	const (
		dbName   = "test_user_list"
		fromTime = 1539244443000
		toTime   = 1539245066000
		timezone = -1
	)

	tldb, err := newTestTradeLogPostgresql(dbName)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, tldb.tearDown(dbName))
	}()
	err = loadTestData(tldb.db, testDataFile)
	require.NoError(t, err)

	from := timeutil.TimestampMsToTime(fromTime)
	to := timeutil.TimestampMsToTime(toTime)

	users, err := tldb.GetUserList(from, to, timezone)
	require.NoError(t, err)
	require.Contains(t, users, common.UserInfo{
		Addr:      "0x8fA07F46353A2B17E92645592a94a0Fc1CEb783F",
		ETHVolume: 0.0022361552369382478,
		USDVolume: 0.5046152992532744,
	})
}