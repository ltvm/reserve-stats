// Code generated by "stringer -type=CEXName -linecomment"; DO NOT EDIT.

package common

import "strconv"

const _CEXName_name = "binancehuobi"

var _CEXName_index = [...]uint8{0, 7, 12}

func (i CEXName) String() string {
	if i < 0 || i >= CEXName(len(_CEXName_index)-1) {
		return "CEXName(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _CEXName_name[_CEXName_index[i]:_CEXName_index[i+1]]
}