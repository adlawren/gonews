package timestamp

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Timestamp contains the data associated with a timestamp stored in the
// database
type Timestamp struct {
	gorm.Model
	T time.Time
}
