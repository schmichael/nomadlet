package uuid

import (
	"crypto/rand"
	"fmt"
)

// Generate is used to generate a random UUID.
func Generate() string {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		panic(fmt.Errorf("go promised this couldn't happen: https://pkg.go.dev/crypto/rand@go1.24.3#Read - %w", err))
	}

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		buf[0:4],
		buf[4:6],
		buf[6:8],
		buf[8:10],
		buf[10:16])
}
