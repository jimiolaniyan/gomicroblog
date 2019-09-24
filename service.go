package gomicroblog

import "fmt"

type service struct {
	users Repository
}

func (svc *service) RegisterNewUser(username string, password string, email string) (*user, error) {
	user, err := NewUser(username, password, email)
	if err != nil {
		return nil, err
	}

	err = svc.users.Store(user)

	if err != nil {
		return nil, fmt.Errorf("error saving user: %s ", err)
	}

	return user, nil
}
