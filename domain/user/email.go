package user

import (
	"errors"
	"net/mail"
	"strings"

	"golang.org/x/net/publicsuffix"
)

var ErrInvalidEmail = errors.New("Invalid Email")

type Email string

func NewEmail(email string) (Email, error) {
	if !isEmail(email) {
		return "", ErrInvalidEmail
	}

	return Email(email), nil
}

func isEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" || len(email) > 254 {
		return false
	}

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return false
	}

	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 {
		return false
	}
	local := email[:at]
	domain := email[at+1:]

	if !isValidLocalPart(local) {
		return false
	}

	if !isValidDomain(domain) {
		return false
	}

	return true
}

func isValidLocalPart(local string) bool {
	if len(local) == 0 || len(local) > 64 {
		return false
	}
	if strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") {
		return false
	}
	if strings.Contains(local, "..") {
		return false
	}

	for _, r := range local {
		if !isAtext(r) && r != '.' {
			return false
		}
	}
	return true
}

func isAtext(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
		return true
	case r >= 'A' && r <= 'Z':
		return true
	case r >= '0' && r <= '9':
		return true
	case strings.ContainsRune("!#$%&'*+-/=?^_`{|}~", r):
		return true
	}
	return false
}

func isValidDomain(domain string) bool {
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}
	if strings.HasPrefix(domain, "-") {
		return false
	}
	if !strings.Contains(domain, ".") {
		return false
	}

	labels := strings.SplitSeq(domain, ".")
	for label := range labels {
		if !isValidLabel(label) {
			return false
		}
	}

	_, isICANN := publicsuffix.PublicSuffix(strings.ToLower(domain))
	return isICANN
}

func isValidLabel(label string) bool {
	if len(label) == 0 || len(label) > 63 {
		return false
	}
	if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
		return false
	}
	for _, r := range label {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return true
}

func (e Email) String() string {
	return string(e)
}

func (e Email) Domain() string {
	email := e.String()

	i := strings.LastIndex(email, "@")
	if i == -1 {
		return ""
	}

	return email[i+1:]
}
