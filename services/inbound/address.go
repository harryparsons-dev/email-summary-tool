// Package inbound generates project email addresses and stores the mail
// delivered to them.
package inbound

import (
	"crypto/rand"
	"fmt"
	"strings"
)

// localPartAlphabet avoids characters that are easy to misread (0/o, 1/l).
const localPartAlphabet = "abcdefghijkmnpqrstuvwxyz23456789"

const localPartLength = 12

// GenerateAddress returns a new random address such as "k3x9p2qahm7d@domain".
func GenerateAddress(domain string) (string, error) {
	buf := make([]byte, localPartLength)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate email address: %w", err)
	}
	for i, b := range buf {
		// 256 is a multiple of the 32-character alphabet, so this is unbiased.
		buf[i] = localPartAlphabet[int(b)%len(localPartAlphabet)]
	}
	return string(buf) + "@" + strings.ToLower(domain), nil
}

// NormalizeAddress lowercases an address and strips surrounding whitespace and
// angle brackets so it can be compared with stored project addresses.
func NormalizeAddress(address string) string {
	address = strings.TrimSpace(address)
	address = strings.TrimPrefix(address, "<")
	address = strings.TrimSuffix(address, ">")
	return strings.ToLower(address)
}

// AddressDomain returns the part of the address after the last "@".
func AddressDomain(address string) string {
	at := strings.LastIndex(address, "@")
	if at < 0 {
		return ""
	}
	return address[at+1:]
}
