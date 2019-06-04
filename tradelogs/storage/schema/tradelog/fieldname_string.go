// Code generated by "stringer -type=FieldName -linecomment"; DO NOT EDIT.

package tradelog

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Time-0]
	_ = x[BlockNumber-1]
	_ = x[TxHash-2]
	_ = x[UserAddr-3]
	_ = x[SrcAddr-4]
	_ = x[DstAddr-5]
	_ = x[Country-6]
	_ = x[IP-7]
	_ = x[EthUSDProvider-8]
	_ = x[DstReserveAddr-9]
	_ = x[SrcReserveAddr-10]
	_ = x[SrcAmount-11]
	_ = x[DstAmount-12]
	_ = x[EthUSDRate-13]
	_ = x[EthAmount-14]
	_ = x[FiatAmount-15]
	_ = x[IntegrationApp-16]
	_ = x[WalletAddress-17]
	_ = x[LogIndex-18]
	_ = x[SourceBurnAmount-19]
	_ = x[DestBurnAmount-20]
	_ = x[SourceWalletFeeAmount-21]
	_ = x[DestWalletFeeAmount-22]
	_ = x[UID-23]
	_ = x[TxSender-24]
}

const _FieldName_name = "timeblock_numbertx_hashuser_addrsrc_addrdst_addrcountryipeth_rate_providerdst_rsv_addrsrc_rsv_addrsrc_amountdst_amounteth_usd_rateeth_amountfiat_amountintegration_appwallet_addrlog_indexsrc_burn_amountdst_burn_amountsrc_wallet_fee_amountdst_wallet_fee_amountuidtx_sender"

var _FieldName_index = [...]uint16{0, 4, 16, 23, 32, 40, 48, 55, 57, 74, 86, 98, 108, 118, 130, 140, 151, 166, 177, 186, 201, 216, 237, 258, 261, 270}

func (i FieldName) String() string {
	if i < 0 || i >= FieldName(len(_FieldName_index)-1) {
		return "FieldName(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _FieldName_name[_FieldName_index[i]:_FieldName_index[i+1]]
}
