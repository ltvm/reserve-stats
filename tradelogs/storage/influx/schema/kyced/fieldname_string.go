// Code generated by "stringer -type=FieldName -linecomment"; DO NOT EDIT.

package kyced

import "strconv"

const _FieldName_name = "timeuser_addrcountrykycedwallet_addr"

var _FieldName_index = [...]uint8{0, 4, 13, 20, 25, 36}

func (i FieldName) String() string {
	if i < 0 || i >= FieldName(len(_FieldName_index)-1) {
		return "FieldName(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _FieldName_name[_FieldName_index[i]:_FieldName_index[i+1]]
}
