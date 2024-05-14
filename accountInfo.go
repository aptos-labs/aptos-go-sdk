package aptos

import (
	"encoding/hex"
	"strconv"
	"strings"
)

// AccountInfo is returned from calls to #Account()
type AccountInfo struct {
	SequenceNumberStr    string `json:"sequence_number"`
	AuthenticationKeyHex string `json:"authentication_key"`
}

// AuthenticationKey Hex decode of AuthenticationKeyHex
func (ai AccountInfo) AuthenticationKey() ([]byte, error) {
	ak := ai.AuthenticationKeyHex
	if strings.HasPrefix(ak, "0x") {
		ak = ak[2:]
	}
	return hex.DecodeString(ak)
}

// SequenceNumber ParseUint of SequenceNumberStr
func (ai AccountInfo) SequenceNumber() (uint64, error) {
	return strconv.ParseUint(ai.SequenceNumberStr, 10, 64)
}
