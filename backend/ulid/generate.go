package ulid

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

// Generate generates a new ULID string.
func Generate() string {
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	res := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	return res.String()
}
