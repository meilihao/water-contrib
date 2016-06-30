package render

import (
	"time"
)

func DateFormat(t time.Time, layout string) string {
	return t.Format(layout)
}
