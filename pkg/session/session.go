package session

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/griffin/cs622-datasec/pkg/user"
	util "github.com/griffin/cs622-datasec/pkg/util"

	"golang.org/x/crypto/bcrypt"
)

const (
	createSession         = "INSERT INTO sessions (validator, selector, user_id, exp) VALUES ($1, $2, $3, $4)"
	selectSession         = "SELECT users.id, users.selector, users.email, users.name, users.postgres_user, sessions.validator, sessions.exp FROM sessions JOIN users ON sessions.user_id=users.id WHERE sessions.selector=$1 "
	validateUser          = "SELECT id, validator FROM users WHERE email=$1"
	invalidateSession     = "DELETE FROM sessions WHERE selector=$1"
	invalidateAllSession  = "DELETE FROM sessions WHERE user_id=$1"
	getAllSessionsForUser = "SELECT selector FROM sessions WHERE user_id=$1"

	selectorLen  = 12
	validatorLen = 20
)

type SessionDatastore interface {
	Create(username, password string, duration time.Duration) (string, error)
	Check(token string) (user.User, error)
	Invalidate(token string) error
	InvalidateAll(usr user.User) error
}

type sessionDatastore struct {
	sqlClient *sql.DB
}

func NewSessionDatastoreHandler(db *sql.DB) SessionDatastore {
	return &sessionDatastore{
		sqlClient: db,
	}
}

func (d *sessionDatastore) Create(email, password string, duration time.Duration) (string, error) {
	var validator string
	var userID uint

	err := d.sqlClient.QueryRow(validateUser, email).Scan(&userID, &validator)
	if err != nil {
		return "", fmt.Errorf("couldn't find user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(validator), []byte(password))
	if err != nil {
		return "", fmt.Errorf("incorrect password: %w", err)
	}

	token, err := d.insertSession(userID, duration)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (d *sessionDatastore) insertSession(userID uint, duration time.Duration) (string, error) {
	validator := util.GenerateSelector(validatorLen)
	selector := util.GenerateSelector(selectorLen)
	exp := time.Now().Add(duration)

	hashValidator := sha256.Sum256([]byte(validator))
	hashValStr := base64.StdEncoding.EncodeToString(hashValidator[:])

	_, err := d.sqlClient.Exec(createSession, hashValStr, selector, userID, exp)
	if err != nil {
		return "", fmt.Errorf("Could not insert session: %v", err)
	}

	return fmt.Sprintf("%v:%v", selector, validator), nil
}

func (d *sessionDatastore) Check(token string) (user.User, error) {
	split := strings.Split(token, ":")
	selector := split[0]
	validator := sha256.Sum256([]byte(split[1]))
	var valQuery string
	var exp time.Time

	var usr user.User

	err := d.sqlClient.QueryRow(selectSession, selector).Scan(&usr.ID, &usr.Selector, &usr.Email, &usr.Name, &usr.PostgresUser, &valQuery, &exp)
	if err != nil {
		return user.User{}, fmt.Errorf("no validator found: %w", err)
	}

	q, err := base64.StdEncoding.DecodeString(valQuery)
	if err != nil {
		return user.User{}, err
	}

	if !bytes.Equal(validator[:], q) {
		return user.User{}, fmt.Errorf("validator != valQ")
	}

	if exp.After(time.Now()) {
		return user.User{}, fmt.Errorf("expired session: now %v, exp %v", time.Now(), exp)
	}

	return usr, nil
}

func (d *sessionDatastore) Invalidate(token string) error {
	split := strings.Split(token, ":")
	selector := split[0]

	_, err := d.sqlClient.Exec(invalidateSession, selector)
	if err != nil {
		return errors.New("couldn't invalidate sessions")
	}

	return nil
}

func (d *sessionDatastore) InvalidateAll(usr user.User) error {
	var v []string

	rows, err := d.sqlClient.Query(getAllSessionsForUser, usr.ID)
	for rows.Next() {
		var validator string
		rows.Scan(&validator)
		v = append(v, validator)
	}

	_, err = d.sqlClient.Exec(invalidateAllSession, usr.ID)
	if err != nil {
		return errors.New("couldn't invalidate all sessions")
	}

	return nil
}
