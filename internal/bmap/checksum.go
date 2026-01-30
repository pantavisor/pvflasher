package bmap

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
)

// GetHasher returns a new hash.Hash for the given algorithm name
func GetHasher(algo string) (hash.Hash, error) {
	switch algo {
	case "sha1":
		return sha1.New(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("unsupported checksum algorithm: %s", algo)
	}
}
