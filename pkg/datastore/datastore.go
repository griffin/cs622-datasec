package datastore

import (
	"database/sql"
	"fmt"

	"github.com/griffin/cs622-datasec/pkg/session"
	"github.com/griffin/cs622-datasec/pkg/user"
	_ "github.com/lib/pq" // Postgres driver
)

type Datastore struct {
	Session session.SessionDatastore
	User    user.UserDatastore

	sqlClient *sql.DB
}

type SQLOptions struct {
	User     string
	Password string
	Host     string
	Port     uint
	Database string
}

func (s SQLOptions) String() string {
	return fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", s.User, s.Password, s.Host, s.Port, s.Database)
}

func New(sqlOpt SQLOptions) (*Datastore, error) {
	d, err := sql.Open("postgres", sqlOpt.String())
	if err != nil {
		return nil, err
	}

	return &Datastore{
		Session: nil,
		User:    nil,

		sqlClient: d,
	}, nil
}
