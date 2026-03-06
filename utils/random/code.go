package random

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

func Code(length int) string {
	if length <= 0 {
		length = 4
	}
	if length > 8 {
		length = 8
	}
	rand.Seed(time.Now().UnixNano())
	max := int(math.Pow10(length))
	return fmt.Sprintf("%0*d", length, rand.Intn(max))
}
