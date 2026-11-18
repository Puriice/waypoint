package unit

import (
	"encoding/json"
	"fmt"
	"time"
)

type Seconds time.Duration

func (s Seconds) Duration() time.Duration {
	return time.Duration(s)
}

func (s Seconds) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(s).Seconds())
}

func (s *Seconds) UnmarshalJSON(data []byte) error {
	var secs float64
	if err := json.Unmarshal(data, &secs); err != nil {
		return fmt.Errorf("invalid seconds: %w", err)
	}
	*s = Seconds(time.Duration(secs * float64(time.Second)))
	return nil
}
