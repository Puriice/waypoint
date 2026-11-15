package domain

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
