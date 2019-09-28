package gomicroblog

import (
	"fmt"
)

type Service interface {
	RegisterNewUser(rm registerUserRequest, res Responder) (*user, error)
}

type service struct {
	users Repository
}

type registerUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (svc *service) RegisterNewUser(username string, password string, email string) (*user, error) {
	user, err := NewUser(username, email)
	if err != nil {
		return nil, err
	}

	if len(password) < 8 {
		return nil, ErrInvalidPassword
	}

	if _, err := verifyNotInUse(svc, username, email); err != nil {
		return nil, err
	}

	user.ID = nextID()
	if hash, err := hashPassword(password); err == nil {
		user.password = hash
	}

	if err = svc.users.Store(user); err != nil {
		return nil, fmt.Errorf("error saving user: %s ", err)
	}

	return user, nil
}

func verifyNotInUse(svc *service, username string, email string) (*user, error) {
	if u, _ := svc.users.FindByName(username); u != nil {
		return nil, fmt.Errorf("username in use")
	}
	if u, _ := svc.users.FindByEmail(email); u != nil {
		return nil, fmt.Errorf("email in use")
	}
	return nil, nil
}
