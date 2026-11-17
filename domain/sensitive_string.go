package domain

import "encoding/json"

type SensitiveString string

func NewSensitiveString(input string) SensitiveString {
	return SensitiveString(input)
}

func (s SensitiveString) Expose() string {
	return string(s)
}

func (SensitiveString) String() string {
	return "[REDACTED]"
}

// MarshalJSON exposes the raw value. String() stays redacted so the value is
// not leaked by fmt or logging, but anything JSON-encoded carries it in full.
func (s SensitiveString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *SensitiveString) UnmarshalJSON(data []byte) error {
	var raw string

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*s = SensitiveString(raw)

	return nil
}
