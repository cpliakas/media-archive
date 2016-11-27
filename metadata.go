package main

import (
	"crypto/md5"
	"encoding/hex"
)

const NumBytes int = 2

// firstBytes returns the first NumBytes bytes in the array.
func firstBytes(dat []byte) []byte {
	var end int
	if end = len(dat); end > NumBytes {
		end = NumBytes
	}
	return dat[:end]
}

// lastBytes returns the last NumBytes bytes in the array.
func lastBytes(dat []byte) []byte {
	var start, end int
	if end = len(dat); end > NumBytes {
		start = end - NumBytes
	}
	return dat[start:end]
}

// hash returns an MD5 hash of the byte array.
func hash(dat []byte) string {
	hasher := md5.New()
	hasher.Write(dat)
	return hex.EncodeToString(hasher.Sum(nil))
}
