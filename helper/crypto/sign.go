package crypto

import (
	"cmp"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"slices"
)

// Param is one signed form parameter.
//
// The value is either a string or an integer, mirroring the Python type
// `tuple[str, str | int]`.
type Param struct {
	Key   string
	Value any
}

// ComputeSign computes the Tieba client signature of data.
//
// It mirrors aiotieba.helper.crypto.sign.compute_sign: the parameters are
// sorted by key, joined as `key=value` without a separator, and the salt is
// appended before taking the MD5 hex digest.
func ComputeSign(data []Param, salt []byte) string {
	sorted := slices.Clone(data)
	slices.SortFunc(sorted, func(a, b Param) int { return cmp.Compare(a.Key, b.Key) })

	h := md5.New()
	for _, p := range sorted {
		fmt.Fprintf(h, "%s=%v", p.Key, p.Value)
	}
	h.Write(salt)
	return hex.EncodeToString(h.Sum(nil))
}

// Sign appends the signature produced by ComputeSign to data.
//
// It mirrors aiotieba.helper.crypto.sign.sign.
func Sign(data []Param, salt []byte) []Param {
	return append(data, Param{Key: "sign", Value: ComputeSign(data, salt)})
}
