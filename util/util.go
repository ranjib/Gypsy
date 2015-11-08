package util

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
)

func UUID() (string, error) {
	u := make([]byte, 16)
	_, err := rand.Read(u)
	if err != nil {
		return "", err
	}
	u[8] = (u[8] | 0x80) & 0xBF
	u[6] = (u[6] | 0x40) & 0x4F
	return hex.EncodeToString(u), nil
}

func Itob(i uint64) []byte {
	v := make([]byte, 8)
	binary.BigEndian.PutUint64(v, i)
	return v
}

func Btoi(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}
