package auth

import (
	"fmt"
	"time"
)

type service struct {
	accounts Repository
	events   Events
}

func NewService(accounts Repository, subscriber Events) Service {
	return &service{accounts: accounts, events: subscriber}
}

func (svc *service) RegisterAccount(r registerAccountRequest) (ID, error) {
	username := r.Username
	email := r.Email

	acc, err := NewAccount(username, email)
	if err != nil {
		return "", err
	}

	password := r.Password
	if len(password) < 8 {
		return "", ErrInvalidPassword
	}

	if _, err := svc.verifyNotInUse(username, email); err != nil {
		return "", err
	}

	acc.ID = NewID()
	if hash, err := hashPassword(password); err == nil {
		acc.Credentials.Password = hash
	}

	acc.CreatedAt = time.Now().UTC()
	if err = svc.accounts.Store(acc); err != nil {
		return "", fmt.Errorf("error saving user: %s ", err)
	}

	svc.events.AccountCreated(string(acc.ID), acc.Credentials.Username, acc.Credentials.Email)

	return acc.ID, nil
}

func (svc *service) verifyNotInUse(username string, email string) (*Account, error) {
	if u, err := svc.accounts.FindByName(username); u != nil && err == nil {
		return nil, ErrExistingUsername
	}

	if u, err := svc.accounts.FindByEmail(email); u != nil && err == nil {
		return nil, ErrExistingEmail
	}

	return nil, nil
}
