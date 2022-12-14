package tradelog

// FieldName define a list of field names for a TradeLog record
//go:generate stringer -type=FieldName -linecomment
type FieldName int

const (
	//Time is enumerated field name for reserveRate.time
	Time FieldName = iota //time
	//BlockNumber is enumerated field name for TradeLog.BlockNumbers
	BlockNumber //block_number
	//TxHash is enumerated field name for tradeLog.TxHash
	TxHash //tx_hash
	//UserAddr is enumerated field name for TradeLog.UserAddr
	UserAddr //user_addr
	//SrcAddr is enumerated field name for TradeLog.SrcAddr
	SrcAddr //src_addr
	//DstAddr is enumerated field name for TradeLog.DstAddr
	DstAddr //dst_addr
	//Country is enumerated field name for TradeLog.CountryName
	Country //country
	//IP is enumerated field name for TradeLog.IP
	IP //ip
	//EthUSDProvider is the enumerated field name for TradeLog.ETHUSDProvider
	EthUSDProvider //eth_rate_provider
	//DstReserveAddr is enumerated fieldname for destination reserve Address
	DstReserveAddr // dst_rsv_addr
	//SrcReserveAddr is enumerated field for source reserve Address
	SrcReserveAddr // src_rsv_addr
	//SrcAmount is the enumerated field for source amount
	SrcAmount //src_amount
	//DstAmount is the enumerated field for source amount
	DstAmount //dst_amount
	//EthUSDRate is the enumerated field for ETH-USD rate
	EthUSDRate // eth_usd_rate
	//EthAmount is the enumerated field for ETH Amount
	EthAmount // eth_amount
	//OriginalEthAmount is the enumerated field for Swap Volume
	OriginalEthAmount // original_eth_amount
	//FiatAmount is the enumerated field for fiat amount
	FiatAmount // fiat_amount
	//IntegrationApp is the name of apps integrated kyberswap
	IntegrationApp //integration_app
	//WalletAddress is the address of wallet associated with trade log
	WalletAddress //wallet_addr
	// WalletName is the name of wallet
	WalletName //wallet_name
	//LogIndex is the index of the log in that block
	LogIndex //log_index
	//SourceBurnAmount is the name of burnFee amount  for source rsv
	SourceBurnAmount //src_burn_amount
	//DestBurnAmount is the name of burnFee amount for dst rsv
	DestBurnAmount //dst_burn_amount
	//SourceWalletFeeAmount is the name of wallet fee Amount for source rsv
	SourceWalletFeeAmount //src_wallet_fee_amount
	//DestWalletFeeAmount is the name of dest wallet fee Amount for dest rsv
	DestWalletFeeAmount //dst_wallet_fee_amount
	// UID is id of KyberSWAP user.
	UID //uid
	// TxSender is address of tx sender
	TxSender //tx_sender
	//ReceiverAddress
	ReceiverAddress //receiver_addr
)

//tradeLogSchemaFields translates the stringer of reserveRate fields into its enumerated form
var tradeLogSchemaFields = map[string]FieldName{
	"time":                  Time,
	"block_number":          BlockNumber,
	"tx_hash":               TxHash,
	"user_addr":             UserAddr,
	"src_addr":              SrcAddr,
	"dst_addr":              DstAddr,
	"country":               Country,
	"ip":                    IP,
	"eth_rate_provider":     EthUSDProvider,
	"dst_rsv_addr":          DstReserveAddr,
	"src_rsv_addr":          SrcReserveAddr,
	"src_amount":            SrcAmount,
	"dst_amount":            DstAmount,
	"eth_usd_rate":          EthUSDRate,
	"eth_amount":            EthAmount,
	"original_eth_amount":   OriginalEthAmount,
	"fiat_amount":           FiatAmount,
	"log_index":             LogIndex,
	"integration_app":       IntegrationApp,
	"wallet_addr":           WalletAddress,
	"wallet_name":           WalletName,
	"src_burn_amount":       SourceBurnAmount,
	"dst_burn_amount":       DestBurnAmount,
	"src_wallet_fee_amount": SourceWalletFeeAmount,
	"dst_wallet_fee_amount": DestWalletFeeAmount,
	"uid":                   UID,
	"tx_sender":             TxSender,
	"receiver_addr":         ReceiverAddress,
}
