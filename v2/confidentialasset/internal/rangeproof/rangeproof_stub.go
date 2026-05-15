//go:build !cgo

package rangeproof

import "errors"

func BatchRangeProof([]uint64, []byte, []byte, []byte, int) ([]byte, []byte, error) {
	return nil, nil, errors.New("rangeproof: CGO disabled")
}
