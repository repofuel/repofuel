package identifier

import (
	"encoding/hex"
)

type Hash [20]byte

func (h Hash) String() string {
	return h.Hex()
}

func (h Hash) ShortHash() string {
	return h.Hex()[:7]
}

var emptyHash Hash

func (h Hash) IsZero() bool {
	return h == emptyHash
}

func (h Hash) Hex() string {
	return hex.EncodeToString(h[:])
}

func (h Hash) Bytes() []byte {
	return h[:]
}

func NewHash(s string) Hash {
	b, _ := hex.DecodeString(s)

	var h Hash
	copy(h[:], b)

	return h
}
