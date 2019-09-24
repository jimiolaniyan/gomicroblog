package gomicroblog

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExistingUsernameCannotBeReused(t *testing.T) {
	svc := service{users: NewUserRepository()}
	user1, _ := svc.RegisterNewUser("username", "password", "a@b")
	_ = svc.users.Store(user1)

	user2, err := svc.RegisterNewUser("username", "password1", "b@c")

	fmt.Println(err)

	assert.Nil(t, user2)
	assert.NotNil(t, err)
}

func TestExistingEmailCannotBeReused(t *testing.T) {
	svc := service{users: NewUserRepository()}
	user1, _ := svc.RegisterNewUser("username", "password", "a@b")
	_ = svc.users.Store(user1)

	user2, err := svc.RegisterNewUser("username2", "password1", "a@b")

	fmt.Println(err)

	assert.Nil(t, user2)
	assert.NotNil(t, err)
}
