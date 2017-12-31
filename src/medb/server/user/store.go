package user

import (
	"encoding/csv"
	"errors"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type userStoreImpl struct {
	// Path to a file that stores the user entries. These should be stored in CSV format:
	// username,passwordHash,pathToDB
	userFilePath string
}

var _ Store = userStoreImpl{}

func (s userStoreImpl) Login(username string, password string) (User, error) {
	f, err := os.Open(s.userFilePath)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	userIndex := -1
	for i, record := range records {
		if record[0] == username {
			// Found our match
			userIndex = i
			break
		}
	}
	if userIndex == -1 {
		return nil, errors.New("user not found")
	}

	// Now see if the password matches
	err = bcrypt.CompareHashAndPassword([]byte(records[userIndex][1]), []byte(password))
	if err != nil {
		return nil, err
	}
	return userImpl{
		username: username,
		path:     records[userIndex][2],
	}, nil
}

type userImpl struct {
	username string
	path     string
}

var _ User = userImpl{}

func (u userImpl) Name() string {
	return u.username
}

func (u userImpl) Path() string {
	return u.path
}
