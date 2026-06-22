package verificationcode

import (
	"fmt"
	"math/rand"
)

func GenerateVerificationCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
