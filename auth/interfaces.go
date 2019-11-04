package auth

type Service interface {
	RegisterAccount(r registerAccountRequest) (ID, error)
}

type Events interface {
	AccountCreated(id string, username string, email string)
}

type Repository interface {
	FindByID(id ID) (*Account, error)
	FindByName(username string) (*Account, error)
	FindByEmail(email string) (*Account, error)
	Store(acc *Account) error
}

type registerAccountRequest struct {
	Username, Email, Password string
}
