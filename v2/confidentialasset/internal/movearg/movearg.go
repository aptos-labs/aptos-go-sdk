// Package movearg builds fully BCS-encoded Move entry-function argument blobs.
//
// All functions return [aptos.RawArg] so that the SDK's writeArg serializer writes
// the bytes verbatim (via ser.FixedBytes) instead of adding an extra ULEB128 length
// prefix. Returning []byte instead would cause double-encoding because writeArg calls
// ser.WriteBytes(v) which prepends its own ULEB128 length.
package movearg

import (
	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset/internal/sigbcs"
)

// VectorU8 encodes a Move vector<u8> as ULEB128(len) || bytes.
func VectorU8(data []byte) aptos.RawArg {
	return aptos.RawArg(sigbcs.AppendBytes(nil, data))
}

// VectorVectorU8 encodes a Move vector<vector<u8>> (e.g. sigma commitments).
func VectorVectorU8(rows [][]byte) aptos.RawArg {
	out := sigbcs.AppendULEB128(nil, uint32(len(rows)))
	for _, r := range rows {
		out = sigbcs.AppendBytes(out, r)
	}
	return aptos.RawArg(out)
}

// VectorTripleVecU8 encodes a Move vector<vector<vector<u8>>> (e.g. voluntary auditor D points).
func VectorTripleVecU8(parts [][][]byte) aptos.RawArg {
	out := sigbcs.AppendULEB128(nil, uint32(len(parts)))
	for _, mid := range parts {
		out = append(out, VectorVectorU8(mid)...)
	}
	return aptos.RawArg(out)
}
