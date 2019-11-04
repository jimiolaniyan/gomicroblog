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
	if u, ok := repo.accounts[id]; ok {
		return u, nil
	}
	return nil, ErrNotFound
}

func (repo *accountRepository) FindByName(username string) (*Account, error) {
	for _, v := range repo.accounts {
		if v.Credentials.Username == username {
			return v, nil
		}
	}
	return nil, ErrNotFound
}

func (repo *accountRepository) FindByEmail(email string) (*Account, error) {
	for _, v := range repo.accounts {
		if v.Credentials.Email == email {
			return v, nil
		}
	}
	return nil, ErrNotFound
}
