package gomicroblog

import (
	"fmt"
)

type Service interface {
	RegisterNewUser(req registerUserRequest) (ID, error)
}

type service struct {
	users Repository
}

type registerUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (svc *service) RegisterNewUser(req registerUserRequest) (ID, error) {
	u := req.Username
	e := req.Email
	user, err := NewUser(u, e)
	if err != nil {
		return "", err
	}

	p := req.Password
	if len(p) < 8 {
		return "", ErrInvalidPassword
	}

	if _, err := verifyNotInUse(svc, u, e); err != nil {
		return "", err
	}

	user.ID = nextID()
	if hash, err := hashPassword(p); err == nil {
		user.password = hash
	}

	if err = svc.users.Store(user); err != nil {
		return "", fmt.Errorf("error saving user: %s ", err)
	}

	return user.ID, nil
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
