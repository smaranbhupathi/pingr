package inbound

import (
	"fmt"
	"net/url"
	"strings"
)

// Validate checks RegisterInput fields and returns a map of field → error message.
// Empty map means valid.
func (i RegisterInput) Validate() map[string]string {
	errs := map[string]string{}

	i.Email = strings.TrimSpace(i.Email)
	i.Username = strings.TrimSpace(i.Username)

	if i.Email == "" {
		errs["email"] = "required"
	} else if !strings.Contains(i.Email, "@") || !strings.Contains(i.Email, ".") {
		errs["email"] = "invalid email address"
	}

	if i.Username == "" {
		errs["username"] = "required"
	} else if len(i.Username) < 3 {
		errs["username"] = "must be at least 3 characters"
	} else if len(i.Username) > 30 {
		errs["username"] = "must be at most 30 characters"
	} else if !isAlphanumeric(i.Username) {
		errs["username"] = "only letters, numbers, and underscores allowed"
	}

	if i.Password == "" {
		errs["password"] = "required"
	} else if len(i.Password) < 8 {
		errs["password"] = "must be at least 8 characters"
	}

	return errs
}

func (i LoginInput) Validate() map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(i.Email) == "" {
		errs["email"] = "required"
	}
	if i.Password == "" {
		errs["password"] = "required"
	}
	return errs
}

func (i CreateMonitorInput) Validate() map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(i.Name) == "" {
		errs["name"] = "required"
	} else if len(i.Name) > 100 {
		errs["name"] = "must be at most 100 characters"
	}

	if strings.TrimSpace(i.URL) == "" {
		errs["url"] = "required"
	} else if err := validateURL(i.URL); err != nil {
		errs["url"] = err.Error()
	}

	allowed := map[int]bool{60: true, 180: true, 300: true}
	if !allowed[i.IntervalSeconds] {
		errs["interval_seconds"] = "must be 60, 180, or 300"
	}

	return errs
}

func validateURL(raw string) error {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("must start with http:// or https://")
	}
	if u.Host == "" {
		return fmt.Errorf("missing host")
	}
	return nil
}

func isAlphanumeric(s string) bool {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}
