package validationx

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

const (
	loginMinLength    = 3
	loginMaxLength    = 30
	passwordMinLength = 8
	passwordMaxLength = 64
)

var reservedLogins = map[string]struct{}{
	"admin": {}, "administrator": {}, "root": {}, "system": {}, "support": {}, "info": {}, "contact": {}, "null": {},
	"undefined": {}, "api": {}, "app": {}, "www": {}, "mail": {}, "email": {},
}

var loginRegexp = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func ValidateLogin(login string) error {
	if len(login) < loginMinLength || len(login) > loginMaxLength {
		return errors.Errorf("login must be %d-%d characters", loginMinLength, loginMaxLength)
	}

	if !loginRegexp.MatchString(login) {
		return errors.New("only latin letters or numbers allowed")
	}

	if _, ok := reservedLogins[strings.ToLower(login)]; ok {
		return errors.New("not allowed login")
	}

	return nil
}

func ValidatePassword(password, login string) error {
	if len(password) < passwordMinLength || len(password) > passwordMaxLength {
		return errors.Errorf("password must be %d-%d characters", passwordMinLength, passwordMaxLength)
	}

	if login != "" && strings.Contains(strings.ToLower(password), strings.ToLower(login)) {
		return errors.New("password must not contain login")
	}

	var (
		hasUpper  = false
		hasLower  = false
		hasNumber = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsSpace(char):
			return errors.New("password cannot contain spaces")
		}
	}

	var missingRequirements []string

	if !hasUpper {
		missingRequirements = append(missingRequirements, "uppercase letter")
	}
	if !hasLower {
		missingRequirements = append(missingRequirements, "lowercase letter")
	}
	if !hasNumber {
		missingRequirements = append(missingRequirements, "digit")
	}

	if len(missingRequirements) > 0 {
		return errors.Errorf("password must contain at least one %s", strings.Join(missingRequirements, ", "))
	}

	return nil
}
