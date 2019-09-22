package gomicroblog

import (
	"fmt"
	"regexp"
)

var emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func Register(username string, password string, email string) (*user, error) {
	if len(username) < 1 {
		return nil, fmt.Errorf("username cannot be empty")
	}

	if len(password) < 8 {
		return nil, fmt.Errorf("invalid password")
	}

	if !emailRegexp.MatchString(email) {
		return nil, fmt.Errorf("invalid email address")
	}

	return &user{}, nil
}
