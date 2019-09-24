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

	if u, _ := svc.users.FindByName(username); u != nil {
		return nil, fmt.Errorf("username in use")
	}

	if u, _ := svc.users.FindByEmail(email); u != nil {
		return nil, fmt.Errorf("email in use")
	}

	if err = svc.users.Store(user); err != nil {
		return nil, fmt.Errorf("error saving user: %s ", err)
	}

	return user, nil
}
