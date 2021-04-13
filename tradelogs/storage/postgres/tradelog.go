package postgres

import (
	"github.com/KyberNetwork/reserve-stats/lib/caller"
	"github.com/KyberNetwork/reserve-stats/lib/pgsql"
	"github.com/KyberNetwork/reserve-stats/tradelogs/common"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/lib/pq"
)

func (tldb *TradeLogDB) saveReserve(reserves []common.Reserve) error {
	var (
		logger    = tldb.sugar.With("func", caller.GetCurrentFunctionName())
		addresses []string
	)
	query := `INSERT INTO reserve (address)
	VALUES(
		UNNEST($1::TEXT[])
	) ON CONFLICT (address) DO NOTHING;`
	logger.Infow("save reserve", "query", query)
	for _, r := range reserves {
		addresses = append(addresses, r.Address.Hex())
	}
	if _, err := tldb.db.Exec(query, pq.StringArray(addresses)); err != nil {
		logger.Errorw("failed to add reserve into db", "error", err)
		return err
	}
	return nil
}

// SaveTradeLogs persist trade logs to DB
func (tldb *TradeLogDB) SaveTradeLogs(crResult *common.CrawlResult) (err error) {
	var (
		logger              = tldb.sugar.With("func", caller.GetCurrentFunctionName())
		reserveAddress      = make(map[string]struct{})
		reserveAddressArray []string
		tokens              = make(map[string]struct{})
		tokensArray         []string
		records             []*record

		users = make(map[ethereum.Address]struct{})
	)
	if crResult != nil {
		if len(crResult.Reserves) > 0 {
			if err := tldb.saveReserve(crResult.Reserves); err != nil {
				return err
			}
		}

		logs := crResult.Trades
		for _, log := range logs {
			r, err := tldb.recordFromTradeLog(log)
			if err != nil {
				return err
			}

			if _, ok := users[log.User.UserAddress]; ok {
				r.IsFirstTrade = false
			} else {
				isFirstTrade, err := tldb.isFirstTrade(log.User.UserAddress)
				if err != nil {
					return err
				}
				r.IsFirstTrade = isFirstTrade
			}
			records = append(records, r)
			users[log.User.UserAddress] = struct{}{}
		}

		for _, r := range records {
			token := r.SrcAddress
			if _, ok := tokens[token]; !ok {
				tokens[token] = struct{}{}
				tokensArray = append(tokensArray, token)
			}
			token = r.DestAddress
			if _, ok := tokens[token]; !ok {
				tokens[token] = struct{}{}
				tokensArray = append(tokensArray, token)
			}
			reserve := r.SrcReserveAddress
			if reserve != "" {
				if _, ok := reserveAddress[reserve]; !ok {
					reserveAddress[reserve] = struct{}{}
					reserveAddressArray = append(reserveAddressArray, reserve)
				}
			}
			reserve = r.DstReserveAddress
			if reserve != "" {
				if _, ok := reserveAddress[reserve]; !ok {
					reserveAddress[reserve] = struct{}{}
					reserveAddressArray = append(reserveAddressArray, reserve)
				}
			}
		}

		tx, err := tldb.db.Beginx()
		if err != nil {
			return err
		}
		defer pgsql.CommitOrRollback(tx, logger, &err)
		if len(reserveAddressArray) > 0 {
			err = tldb.saveReserveAddress(tx, reserveAddressArray)
			if err != nil {
				logger.Debugw("failed to save reserve address", "error", err)
				return err
			}
		}

		err = tldb.saveTokens(tx, tokensArray)
		if err != nil {
			logger.Debugw("failed to save token", "error", err)
			return err
		}

		for _, r := range records {
			logger.Debugw("Record", "record", r)
			_, err = tx.NamedExec(insertionUserTemplate, r)
			if err != nil {
				logger.Infow("user", "address", r.UserAddress)
				logger.Debugw("Error while add users", "error", err)
				return err
			}

			query := `SELECT _id as id FROM 
			create_or_update_tradelogs(
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
				$13, $14, $15, $16, $17, $18, $19, $20
			);`
			var tradelogID uint64
			if err != nil {
				logger.Debugw("failed to prepare fees record", "error", err)
				return err
			}
			if err := tx.Get(&tradelogID, query, 0,
				r.Timestamp, r.BlockNumber, r.TransactionHash,
				r.USDTAmount, r.OriginalUSDTAmount,
				r.ReserveAddress,
				r.UserAddress, r.SrcAddress, r.DestAddress,
				r.SrcAmount, r.DestAmount,
				r.Index,
				r.IsFirstTrade,
				r.TxSender,
				r.ReceiverAddress,
				r.GasUsed,
				r.GasPrice,
				r.TransactionFee,
				r.Version,
			); err != nil {
				logger.Debugw("failed to save tradelogs", "error", err)
				return err
			}
		}

		return err
	}
	return nil
}

func (tldb *TradeLogDB) isFirstTrade(userAddr ethereum.Address) (bool, error) {
	query := `SELECT NOT EXISTS(SELECT NULL FROM "users" WHERE address=$1);`
	row := tldb.db.QueryRow(query, userAddr.Hex())
	var result bool
	if err := row.Scan(&result); err != nil {
		tldb.sugar.Error(err)
		return false, err
	}
	return result, nil
}