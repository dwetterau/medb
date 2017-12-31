package user

type Store interface {
	Login(username string, password string) (User, error)
}

func NewStore(userFilePath string) Store {
	return userStoreImpl{userFilePath}
}

type User interface {
	Path() string
}
