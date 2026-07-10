package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const (
	tokenPrefix   = "drp_"
	tokenRandLen  = 40
	identPrefixLen = 8
	charsetAlpha  = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

// NewAPIToken membuat token baru: "drp_" + 40 char random.
// Mengembalikan (tokenPlain, tokenHash, tokenPrefix8).
func NewAPIToken() (plain, hash, prefix string, err error) {
	b := make([]byte, tokenRandLen)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charsetAlpha))))
		if err != nil {
			return "", "", "", fmt.Errorf("rand: %w", err)
		}
		b[i] = charsetAlpha[n.Int64()]
	}
	plain = tokenPrefix + string(b)
	hash = HashToken(plain)
	prefix = plain[:len(tokenPrefix)+identPrefixLen]
	return plain, hash, prefix, nil
}

// HashToken mengembalikan SHA-256 hex dari token plain.
func HashToken(plain string) string {
	h := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(h[:])
}

// GenerateVoucherCredential menghasilkan string acak sepanjang length
// dari charset alfanumerik (tanpa karakter ambigu: 0/O, 1/I).
func GenerateVoucherCredential(length int) (string, error) {
	if length <= 0 {
		length = 8
	}
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charsetAlpha))))
		if err != nil {
			return "", fmt.Errorf("rand: %w", err)
		}
		b[i] = charsetAlpha[n.Int64()]
	}
	return string(b), nil
}

// ApplyPattern menerapkan pola "####-####" (# diganti char acak).
// Jika pattern kosong, fallback ke random length.
func ApplyPattern(pattern string, length int) (string, error) {
	if pattern == "" {
		return GenerateVoucherCredential(length)
	}
	var sb strings.Builder
	for _, ch := range pattern {
		if ch == '#' {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charsetAlpha))))
			if err != nil {
				return "", fmt.Errorf("rand: %w", err)
			}
			sb.WriteByte(charsetAlpha[n.Int64()])
		} else {
			sb.WriteRune(ch)
		}
	}
	return sb.String(), nil
}

// PrefixedName membuat "<prefix>-<NNN>" dengan zero-pad.
func PrefixedName(prefix string, seq int, pad int) string {
	if pad <= 0 {
		pad = 3
	}
	return fmt.Sprintf("%s-%0*d", prefix, pad, seq)
}

// NowPtr helper untuk pointer time.
func NowPtr() *time.Time {
	t := time.Now()
	return &t
}
