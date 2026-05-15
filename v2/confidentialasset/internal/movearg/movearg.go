// Package movearg builds Move entry-function argument BCS blobs passed as []any to aptos.EntryFunctionPayload
// (serializeArg([]byte) returns bytes verbatim; each logical Move arg must be fully BCS-encoded).
package movearg

import (
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/sigbcs"
)

// VectorU8 is BCS for Move vector<u8> (ULEB128 length + bytes).
func VectorU8(data []byte) []byte {
	return sigbcs.AppendBytes(nil, data)
}

// VectorVectorU8 is BCS for Move vector<vector<u8>> (e.g. sigma commitments).
func VectorVectorU8(rows [][]byte) []byte {
	out := sigbcs.AppendULEB128(nil, uint32(len(rows)))
	for _, r := range rows {
		out = sigbcs.AppendBytes(out, r)
	}
	return out
}

// VectorTripleVecU8 encodes vector<vector<vector<u8>>> (e.g. voluntary auditor D points).
func VectorTripleVecU8(parts [][][]byte) []byte {
	out := sigbcs.AppendULEB128(nil, uint32(len(parts)))
	for _, mid := range parts {
		out = append(out, VectorVectorU8(mid)...)
	}
	return out
}
