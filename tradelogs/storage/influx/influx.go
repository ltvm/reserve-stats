package influx

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/influxdata/influxdb/client/v2"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-stats/lib/blockchain"
	"github.com/KyberNetwork/reserve-stats/lib/caller"
	"github.com/KyberNetwork/reserve-stats/lib/influxdb"
	"github.com/KyberNetwork/reserve-stats/tradelogs/common"
	kycedschema "github.com/KyberNetwork/reserve-stats/tradelogs/storage/influx/schema/kyced"
	logschema "github.com/KyberNetwork/reserve-stats/tradelogs/storage/influx/schema/tradelog"
	"github.com/KyberNetwork/reserve-stats/tradelogs/storage/utils"
)

const (
	//timePrecision is the precision configured for influxDB
	timePrecision           = "s"
	tradeLogMeasurementName = "trades"
)

// Storage represent a client to store trade data to influx DB
type Storage struct {
	sugar                *zap.SugaredLogger
	dbName               string
	influxClient         client.Client
	tokenAmountFormatter blockchain.TokenAmountFormatterInterface

	// traded stored traded addresses to use in a single SaveTradeLogs
	traded      map[ethereum.Address]struct{}
	tokenSymbol sync.Map

	// used for calculate burn amount
	// as different environment have different knc address
	kncAddr ethereum.Address
}

// NewInfluxStorage init an instance of Storage
func NewInfluxStorage(sugar *zap.SugaredLogger, dbName string, influxClient client.Client,
	tokenAmountFormatter blockchain.TokenAmountFormatterInterface, kncAddr ethereum.Address) (*Storage, error) {
	storage := &Storage{
		sugar:                sugar,
		dbName:               dbName,
		influxClient:         influxClient,
		tokenAmountFormatter: tokenAmountFormatter,
		traded:               make(map[ethereum.Address]struct{}),
		tokenSymbol:          sync.Map{},
		kncAddr:              kncAddr,
	}
	storage.tokenSymbol.Store(ethereum.HexToAddress("0x89d24a6b4ccb1b6faa2625fe562bdd9a23260359").Hex(), "SAI")
	if err := storage.createDB(); err != nil {
		return nil, err
	}
	return storage, nil
}

// SaveTradeLogs persist trade logs to DB
func (is *Storage) SaveTradeLogs(logs []common.TradelogV4) error {
	logger := is.sugar.With("func", caller.GetCurrentFunctionName())
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  is.dbName,
		Precision: timePrecision,
	})
	if err != nil {
		return err
	}
	for _, log := range logs {
		points, err := is.tradeLogToPoint(log)
		if err != nil {
			return err
		}

		for _, pt := range points {
			bp.AddPoint(pt)
		}
	}

	if err := is.influxClient.Write(bp); err != nil {
		logger.Errorw("saving error", "error", err)
		return err
	}

	if len(logs) > 0 {
		logger.Debugw("saved trade logs into influxdb",
			"first_block", logs[0].BlockNumber,
			"last_block", logs[len(logs)-1].BlockNumber,
			"trade_logs", len(logs))
	} else {
		logger.Debugw("no trade log to store")
	}

	// reset traded map to avoid ever growing size
	is.traded = make(map[ethereum.Address]struct{})
	return nil
}

// LastBlock returns last stored trade log block number from database.
func (is *Storage) LastBlock() (int64, error) {
	q := `SELECT "block_number","eth_amount" from "trades" ORDER BY time DESC limit 1`

	res, err := influxdb.QueryDB(is.influxClient, q, is.dbName)
	if err != nil {
		return 0, err
	}

	if len(res) != 1 || len(res[0].Series) != 1 || len(res[0].Series[0].Values[0]) != 3 {
		is.sugar.Info("no result returned for last block query")
		return 0, nil
	}

	return influxdb.GetInt64FromInterface(res[0].Series[0].Values[0][1])
}

func prepareTradeLogQuery() string {
	var (
		tradeLogQueryFields = []logschema.FieldName{
			logschema.Time,
			logschema.BlockNumber,
			logschema.OriginalEthAmount,
			logschema.EthAmount,
			logschema.UserAddr,
			logschema.SrcAddr,
			logschema.DstAddr,
			logschema.SrcAmount,
			logschema.DstAmount,
			logschema.IP,
			logschema.Country,
			logschema.UID,
			logschema.IntegrationApp,
			logschema.SourceBurnAmount,
			logschema.DestBurnAmount,
			logschema.LogIndex,
			logschema.TxHash,
			logschema.SrcReserveAddr,
			logschema.DstReserveAddr,
			logschema.SourceWalletFeeAmount,
			logschema.DestWalletFeeAmount,
			logschema.WalletAddress,
			logschema.TxSender,
			logschema.ReceiverAddress,
		}
		tradeLogQuery string
	)
	for _, field := range tradeLogQueryFields {
		tradeLogQuery += field.String() + ", "
	}
	fiatAmount := fmt.Sprintf("(%s * %s) AS %s", logschema.EthAmount.String(), logschema.EthUSDRate.String(), logschema.FiatAmount.String())
	tradeLogQuery += fiatAmount
	return tradeLogQuery
}

// LoadTradeLogsByTxHash get list of tradelogs by tx hash
func (is *Storage) LoadTradeLogsByTxHash(tx ethereum.Hash) ([]common.TradeLog, error) {
	return nil, errors.New("influx does not supported get tradelog by txhash")
}

// LoadTradeLogs return trade logs from DB
func (is *Storage) LoadTradeLogs(from, to time.Time) ([]common.TradeLog, error) {
	var (
		result = make([]common.TradeLog, 0)
		q      = fmt.Sprintf(
			`
		SELECT %[1]s FROM %[2]s WHERE time >= '%[3]s' AND time <= '%[4]s';
		`,
			prepareTradeLogQuery(),
			tradeLogMeasurementName,
			from.Format(time.RFC3339),
			to.Format(time.RFC3339),
		)

		logger = is.sugar.With(
			"func", caller.GetCurrentFunctionName(),
			"from", from,
			"to", to,
		)
	)
	logger.Debug("prepared query statement", "query", q)

	res, err := influxdb.QueryDB(is.influxClient, q, is.dbName)
	if err != nil {
		return nil, err
	}

	// Get TradeLogs
	if len(res) == 0 || len(res[0].Series) == 0 || len(res[0].Series[0].Values) == 0 {
		is.sugar.Debug("empty trades in query result")
		return result, nil
	}
	idxs, err := logschema.NewFieldsRegistrar(res[0].Series[0].Columns)
	if err != nil {
		return nil, err
	}
	for _, row := range res[0].Series[0].Values {

		tradeLog, err := is.rowToTradeLog(row, idxs)
		if err != nil {
			return nil, err
		}
		result = append(result, tradeLog)
	}

	return result, nil
}

// createDB creates the database will be used for storing trade logs measurements.
func (is *Storage) createDB() error {
	_, err := influxdb.QueryDB(is.influxClient, fmt.Sprintf("CREATE DATABASE %s", is.dbName), is.dbName)
	return err
}

// func (is *Storage) getWalletFeeAmount(log common.TradeLog) (float64, float64, error) {
// 	var (
// 		logger = is.sugar.With(
// 			"func", caller.GetCurrentFunctionName(),
// 			"log", log,
// 		)
// 		dstAmount    float64
// 		srcAmount    float64
// 		srcAmountSet bool
// 	)
// 	for _, walletFee := range log.WalletFees {
// 		amount, err := is.tokenAmountFormatter.FromWei(blockchain.KNCAddr, walletFee.Amount)
// 		if err != nil {
// 			return dstAmount, srcAmount, err
// 		}

// 		switch {
// 		case walletFee.ReserveAddress == log.SrcReserveAddress && !srcAmountSet:
// 			srcAmount = amount
// 			// to prevent setting SrcReserveAddress twice when SrcReserveAddress =DstReserveAddress
// 			srcAmountSet = true
// 		case walletFee.ReserveAddress == log.DstReserveAddress:
// 			dstAmount = amount
// 		default:
// 			logger.Warnw("unexpected wallet fees with unrecognized reserve address", "wallet fee", walletFee)
// 		}
// 	}
// 	return srcAmount, dstAmount, nil
// }

func (is *Storage) tradeLogToPoint(log common.TradelogV4) ([]*client.Point, error) {
	var points []*client.Point

	walletAddr := log.WalletAddress
	walletName := log.WalletName

	tags := map[string]string{

		logschema.UserAddr.String(): log.User.UserAddress.String(),

		logschema.SrcAddr.String():        log.TokenInfo.SrcAddress.String(),
		logschema.DstAddr.String():        log.TokenInfo.DestAddress.String(),
		logschema.IntegrationApp.String(): log.IntegrationApp,
		logschema.LogIndex.String():       strconv.FormatUint(uint64(log.Index), 10),

		logschema.Country.String(): log.User.Country,

		logschema.LogIndex.String():        strconv.FormatUint(uint64(log.Index), 10),
		logschema.UID.String():             log.User.UID,
		logschema.TxSender.String():        log.TxDetail.TxSender.String(),
		logschema.ReceiverAddress.String(): log.ReceiverAddress.String(),
	}

	// if !blockchain.IsZeroAddress(log.SrcReserveAddress) {
	// 	tags[logschema.SrcReserveAddr.String()] = log.SrcReserveAddress.String()
	// }
	// if !blockchain.IsZeroAddress(log.DstReserveAddress) {
	// 	tags[logschema.DstReserveAddr.String()] = log.DstReserveAddress.String()
	// }
	if !blockchain.IsZeroAddress(walletAddr) {
		tags[logschema.WalletAddress.String()] = walletAddr.String()
	}

	if len(walletName) != 0 {
		tags[logschema.WalletName.String()] = walletName
	}

	ethAmount, err := is.tokenAmountFormatter.FromWei(blockchain.ETHAddr, log.EthAmount)
	if err != nil {
		return nil, err
	}

	originalEthAmount, err := is.tokenAmountFormatter.FromWei(blockchain.ETHAddr, log.OriginalEthAmount)
	if err != nil {
		return nil, err
	}

	srcAmount, err := is.tokenAmountFormatter.FromWei(log.TokenInfo.SrcAddress, log.SrcAmount)
	if err != nil {
		return nil, err
	}

	dstAmount, err := is.tokenAmountFormatter.FromWei(log.TokenInfo.DestAddress, log.DestAmount)
	if err != nil {
		return nil, err
	}

	srcBurnAmount, dstBurnAmount, err := utils.GetBurnAmount(is.sugar, is.tokenAmountFormatter, log, is.kncAddr)
	if err != nil {
		return nil, err
	}

	// srcWalletFee, dstWalletFee, err := is.getWalletFeeAmount(log)
	// if err != nil {
	// 	return nil, err
	// }
	fields := map[string]interface{}{

		logschema.SrcAmount.String():        srcAmount,
		logschema.DstAmount.String():        dstAmount,
		logschema.EthUSDRate.String():       log.ETHUSDRate,
		logschema.SourceBurnAmount.String(): srcBurnAmount,
		logschema.DestBurnAmount.String():   dstBurnAmount,

		logschema.EthAmount.String():         ethAmount,
		logschema.OriginalEthAmount.String(): originalEthAmount,
		logschema.BlockNumber.String():       int64(log.BlockNumber),
		logschema.TxHash.String():            log.TransactionHash.String(),
		logschema.IP.String():                log.User.IP,
		logschema.EthUSDProvider.String():    log.ETHUSDProvider,
		// logschema.SourceWalletFeeAmount.String(): srcWalletFee,
		// logschema.DestWalletFeeAmount.String():   dstWalletFee,
	}
	tradePoint, err := client.NewPoint(tradeLogMeasurementName, tags, fields, log.Timestamp)
	if err != nil {
		return nil, err
	}

	points = append(points, tradePoint)

	firstTradePoint, err := is.assembleFirstTradePoint(log)
	if err != nil {
		return nil, err
	}
	if firstTradePoint != nil {
		points = append(points, firstTradePoint)
	}

	kycedPoint, err := is.AssembleKYCPoint(log)
	if err != nil {
		return nil, err
	}

	if kycedPoint != nil {
		points = append(points, kycedPoint)
	}

	return points, nil
}

func (is *Storage) assembleFirstTradePoint(logItem common.TradelogV4) (*client.Point, error) {
	var logger = is.sugar.With(
		"func", caller.GetCurrentFunctionName(),
		"timestamp", logItem.Timestamp.String(),
		"user_addr", logItem.User.UserAddress.Hex(),
		"country", logItem.User.Country,
	)

	if _, ok := is.traded[logItem.User.UserAddress]; ok {
		logger.Debug("user has already traded, ignoring")
		return nil, nil
	}

	traded, err := is.userTraded(logItem.User.UserAddress)
	if err != nil {
		return nil, err
	}

	if traded {
		return nil, nil
	}

	logger.Debugw("user first trade")
	tags := map[string]string{
		"user_addr": logItem.User.UserAddress.Hex(),
		"country":   logItem.User.Country,
	}
	tags["wallet_addr"] = logItem.WalletAddress.Hex()

	fields := map[string]interface{}{
		"traded": true,
	}

	point, err := client.NewPoint("first_trades", tags, fields, logItem.Timestamp)
	if err != nil {
		return nil, err
	}

	is.traded[logItem.User.UserAddress] = struct{}{}
	return point, nil
}

func (is *Storage) userTraded(addr ethereum.Address) (bool, error) {
	q := fmt.Sprintf("SELECT traded FROM first_trades WHERE user_addr='%s'", addr.String())
	response, err := influxdb.QueryDB(is.influxClient, q, is.dbName)
	if err != nil {
		return false, err
	}
	// if there is no record, this mean the address has not traded yet
	if (len(response) == 0) || (len(response[0].Series) == 0) || (len(response[0].Series[0].Values) == 0) {
		return false, nil
	}
	return true, nil
}

// AssembleKYCPoint constructs kyced InfluxDB data point from given trade log.
func (is *Storage) AssembleKYCPoint(logItem common.TradelogV4) (*client.Point, error) {
	var (
		logger = is.sugar.With(
			"func", caller.GetCurrentFunctionName(),
			"timestamp", logItem.Timestamp.String(),
			"user_addr", logItem.User.UserAddress.Hex(),
			"country", logItem.User.Country,
		)
		kyced bool
	)

	if logItem.User.UID != "" {
		kyced = true
	}

	if !kyced {
		logger.Debugw("user has not been kyced yet")
		return nil, nil
	}

	logger.Debugw("user has been kyced")
	tags := map[string]string{
		kycedschema.UserAddress.String(): logItem.User.UserAddress.Hex(),
		kycedschema.Country.String():     logItem.User.Country,
	}
	tags[kycedschema.WalletAddress.String()] = logItem.WalletAddress.Hex()

	fields := map[string]interface{}{
		kycedschema.KYCed.String(): true,
	}

	point, err := client.NewPoint("kyced", tags, fields, logItem.Timestamp)
	return point, err
}

// GetStats mock function return StatsResponse
func (is *Storage) GetStats(from, to time.Time) (common.StatsResponse, error) {
	return common.StatsResponse{}, nil
}

// GetTopTokens mock function return TopTokens
func (is *Storage) GetTopTokens(from, to time.Time, limit uint64) (common.TopTokens, error) {
	return common.TopTokens{}, nil
}

// GetTopIntegrations mock function return TopIntegrations
func (is *Storage) GetTopIntegrations(from, to time.Time, limit uint64) (common.TopIntegrations, error) {
	return common.TopIntegrations{}, nil
}

// GetTopReserves mock function return TopReserves
func (is *Storage) GetTopReserves(from, to time.Time, limit uint64) (common.TopReserves, error) {
	return common.TopReserves{}, nil
}

// GetNotTwittedTrades return not twitted trades
func (is *Storage) GetNotTwittedTrades(from, to time.Time) ([]common.BigTradeLog, error) {
	return nil, nil
}

// SaveBigTrades save trades detected big into storage
func (is *Storage) SaveBigTrades(bigVolume float32, fromBlock uint64) error {
	return nil
}

// UpdateBigTradesTwitted update big trade into storage
func (is *Storage) UpdateBigTradesTwitted(trades []uint64) error {
	return nil
}
