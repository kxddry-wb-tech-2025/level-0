package timefmt

import (
	"encoding/json"
	"time"
)

type TimeFmt time.Time

func (t TimeFmt) MarshalJSON() ([]byte, error) {
	formatted := time.Time(t).UTC().Format("2006-01-02T15:04:05Z")
	return json.Marshal(formatted)
}
