// Code generated by "stringer -type=FieldName -linecomment"; DO NOT EDIT.

package heatmap

import "strconv"

const _FieldName_name = "timeeth_volumetoken_volumeusd_volumecountrydst_addrsrc_addr"

var _FieldName_index = [...]uint8{0, 4, 14, 26, 36, 43, 51, 59}

func (i FieldName) String() string {
	if i < 0 || i >= FieldName(len(_FieldName_index)-1) {
		return "FieldName(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _FieldName_name[_FieldName_index[i]:_FieldName_index[i+1]]
}
