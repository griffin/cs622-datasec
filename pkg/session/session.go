package session

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/griffin/cs622-datasec/pkg/user"

	"golang.org/x/crypto/bcrypt"
)

const (
	createSession         = "INSERT INTO sessions (validator, selector, user_id, exp) VALUES ($1, $2, $3, $4)"
	selectSession         = "SELECT users.id, users.selector, users.email, users.first_name, users.last_name, users.gender, users.dob, sessions.validator, sessions.exp FROM sessions JOIN users ON sessions.user_id=users.id WHERE sessions.selector=$1 "
	validateUser          = "SELECT id, selector, validator, first_name, last_name, gender, dob FROM users WHERE email=$1"
	invalidateSession     = "DELETE FROM sessions WHERE selector=$1"
	invalidateAllSession  = "DELETE FROM sessions WHERE user_id=$1"
	getAllSessionsForUser = "SELECT selector FROM sessions WHERE user_id=$1"

	selectorLen  = 12
	validatorLen = 20
)

type SessionDatastore interface {
	CreateSession(username, password string) (user.User, string, error)
	CheckSession(token string) (user.User, error)
	InvalidateSession(token string) error
}

type sessionDatastore struct {
	sqlClient *sql.DB
}

func (d *sessionDatastore) CreateSession(email, password string) (user.User, string, error) {
	usr := &user.User{}
	var validator string

	err := d.sqlClient.QueryRow(validateUser, email).Scan(&usr.ID, &usr.sel, &validator, &usr.FirstName, &usr.LastName, &usr.Gender, &usr.DateOfBirth)
	if err != nil {
		return nil, "", errors.New("couldn't find user")
	}

	err = bcrypt.CompareHashAndPassword([]byte(validator), []byte(password))
	if err != nil {
		return nil, "", errors.New("incorrect password")
	}

	token, err := d.insertSession(usr)
	if err != nil {
		return nil, "", err
	}

	return usr, token, nil
}

func (d *sessionDatastore) insertSession(user user.User) (string, error) {
	validator := d.GenerateSelector(validatorLen)
	selector := d.GenerateSelector(selectorLen)
	exp := time.Now().Add(time.Hour * 2).Unix() //TODO

	hashValidator := sha256.Sum256([]byte(validator))
	hashValStr := base64.StdEncoding.EncodeToString(hashValidator[:])

	_, err := d.sqlClient.Exec(createSession, hashValStr, selector, user.ID, exp)
	if err != nil {
		return "", fmt.Errorf("Could not insert session: %v", err)
	}

	return fmt.Sprintf("%v:%v", selector, validator), nil
}

func (d *sessionDatastore) CheckSession(token string) (user.User, error) {
	split := strings.Split(token, ":")
	selector := split[0]
	validator := sha256.Sum256([]byte(split[1]))
	var valQuery string
	var exp int64

	var usr user.User

	err := d.sqlClient.QueryRow(selectSession, selector).Scan(&usr.ID, &usr.sel, &usr.Email, &usr.FirstName, &usr.LastName, &valQuery, &exp)
	if err != nil {
		return nil, errors.New("no validator found")
	}

	//check:

	q, err := base64.StdEncoding.DecodeString(valQuery)
	if err != nil {
		return user.User{}, err
	}

	if !bytes.Equal(validator[:], q) {
		return user.User{}, fmt.Errorf("validator != valQ")
	}

	if time.Now().Unix() > exp {
		return user.User{}, fmt.Errorf("expired session")
	}

	return usr, nil
}

func (d *sessionDatastore) InvalidateSession(token string) error {
	split := strings.Split(token, ":")
	selector := split[0]

	_, err := d.sqlClient.Exec(invalidateSession, selector)
	if err != nil {
		return errors.New("couldn't invalidate sessions")
	}

	err = d.sqlClient.Del(selector).Err()

	return nil
}

func (d *sessionDatastore) InvalidateAllSessions(usr User) error {
	var v []string

	rows, err := d.sqlClient.Query(getAllSessionsForUser, usr.ID)
	for rows.Next() {
		var validator string
		rows.Scan(&validator)
		v = append(v, validator)
	}

	err = d.Del(v...).Err()

	_, err = d.sqlClient.Exec(invalidateAllSession, usr.ID)
	if err != nil {
		return errors.New("couldn't invalidate all sessions")
	}

	return nil
}
