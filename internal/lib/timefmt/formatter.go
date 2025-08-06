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

func (t *TimeFmt) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	parsed, err := time.Parse("2006-01-02T15:04:05Z", str)
	if err != nil {
		return err
	}

	*t = TimeFmt(parsed)
	return nil
}
