package scanner

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Fingerprint returns a stable short hash of file path, line number, and raw value.
func Fingerprint(file string, line int, value string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%s", file, line, value)))
	return hex.EncodeToString(sum[:])[:16]
}
