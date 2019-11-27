package postgres

import (
	"fmt"
	"time"

	"github.com/KyberNetwork/reserve-stats/lib/caller"
	"github.com/KyberNetwork/reserve-stats/tradelogs/common"
	"github.com/KyberNetwork/reserve-stats/tradelogs/storage/postgres/schema"
)

func (tldb *TradeLogDB) GetUserList(fromTime, toTime time.Time) ([]common.UserInfo, error) {
	logger := tldb.sugar.With("from", fromTime, "to", toTime, "func", caller.GetCurrentFunctionName())

	userListQuery := fmt.Sprintf(`
		SELECT b.address user_address,sum(eth_amount) total_eth_volume,
			sum(eth_amount * eth_usd_rate) total_usd_volume
		FROM "%[1]s" a
		INNER JOIN "%[2]s" b ON a.user_address_id =b.id
		WHERE a.timestamp >= $1 and a.timestamp <= $2
		GROUP BY user_address
	`, schema.TradeLogsTableName, schema.UserTableName)
	logger.Debugw("prepare statement", "stmt", userListQuery)

	var result []common.UserInfo
	if err := tldb.db.Select(&result, userListQuery, fromTime,
		toTime); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		logger.Debugw("empty user list result", "query", userListQuery)
		return nil, nil
	}
	return result, nil
}