package gomicroblog

import (
	"fmt"
)

type Service interface {
	RegisterNewUser(req registerUserRequest, res Responder)
}

type service struct {
	users Repository
}

type registerUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (svc *service) RegisterNewUser(req registerUserRequest) (*user, error) {
	u := req.Username
	e := req.Email
	user, err := NewUser(u, e)
	if err != nil {
		return nil, err
	}

	p := req.Password
	if len(p) < 8 {
		return nil, ErrInvalidPassword
	}

	if _, err := verifyNotInUse(svc, u, e); err != nil {
		return nil, err
	}

	user.ID = nextID()
	if hash, err := hashPassword(p); err == nil {
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
