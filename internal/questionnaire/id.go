package questionnaire

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"
)

func generateID(withTimestamp bool, len int) string {
	randomBytes := make([]byte, len)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}

	plaintext := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	if !withTimestamp {
		return plaintext
	}

	timestamp := time.Now().Unix()
	return fmt.Sprintf("%d%s", timestamp, plaintext)
}
