package user

import (
	"database/sql"
	"errors"

	util "github.com/griffin/cs622-datasec/pkg/util"

	"golang.org/x/crypto/bcrypt"
)

const (
	createUser = "INSERT INTO users (selector, validator, name, email) VALUES ($1, $2, $3, $4) RETURNING id"
	getUser    = "SELECT id, name, email FROM users WHERE selector=$1"
	deleteUser = "DELETE FROM users WHERE id=$1"

	selectorLen = 12
)

type User struct {
	Selector string
	ID       uint

	Email string
	Name  string
}

type userDatastore struct {
	sqlClient *sql.DB
}

type UserDatastore interface {
	Create(usr User, password string) (*User, error)
	Get(string) (*User, error)
	Delete(usr User) error
}

func (d *userDatastore) Create(usr User, password string) (*User, error) {
	validator, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	usr.Selector = util.GenerateSelector(selectorLen)
	if err != nil {
		return nil, err
	}

	err = d.sqlClient.QueryRow(createUser, usr.Selector, string(validator), usr.Name, usr.Email).Scan(&usr.ID)
	if err != nil {
		return nil, err
	}

	return &usr, nil
}

func (d *userDatastore) Get(selector string) (*User, error) {
	var usr User

	err := d.sqlClient.QueryRow(getUser, selector).Scan(&usr.ID, &usr.Name, &usr.Email)
	if err != nil {
		return nil, errors.New("Couldn't find user")
	}

	return &usr, nil
}

func (d *userDatastore) Delete(usr User) error {
	_, err := d.sqlClient.Exec(deleteUser, usr.ID)
	if err != nil {
		return errors.New("delete user failed")
	}

	return nil
}
