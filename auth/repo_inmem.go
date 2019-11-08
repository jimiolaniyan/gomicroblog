package auth

type accountRepository struct {
	accounts map[ID]*Account
}

func NewAccountRepository() Repository {
	return &accountRepository{accounts: map[ID]*Account{}}
}

func (repo *accountRepository) Store(acc *Account) error {
	repo.accounts[acc.ID] = acc
	return nil
}

func (repo *accountRepository) FindByID(id ID) (*Account, error) {
	if a, ok := repo.accounts[id]; ok {
		return a, nil
	}
	return nil, ErrNotFound
}

func (repo *accountRepository) FindByName(username string) (*Account, error) {
	for _, a := range repo.accounts {
		if a.Credentials.Username == username {
			return a, nil
		}
	}
	return nil, ErrNotFound
}

func (repo *accountRepository) FindByEmail(email string) (*Account, error) {
	for _, a := range repo.accounts {
		if a.Credentials.Email == email {
			return a, nil
		}
	}
	return nil, ErrNotFound
}
