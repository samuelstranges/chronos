package util

import (
	"crypto/rand"
	"fmt"

	"github.com/samuelstranges/chronos/types"
)

// GenerateUUID creates an RFC4122-compliant UUID v4
// See: https://icalendar.org/New-Properties-for-iCalendar-RFC-7986/5-3-uid-property.html
func GenerateUUID() (string, error) {
	b := make([]byte, types.HashShift16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Set version (4) and variant bits according to RFC4122
	b[6] = (b[6] & 0x0f) | types.HashMask0x40 // Version 4
	b[8] = (b[8] & 0x3f) | types.HashMask0x80 // Variant bits

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}