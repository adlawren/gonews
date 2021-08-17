package timestamp

import (
	"time"
)

// Timestamp contains the data associated with a timestamp stored in the
// database
type Timestamp struct {
	ID   uint
	T    time.Time
	Name string
}
