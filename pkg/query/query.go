package query

import (
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/griffin/cs622-datasec/pkg/user"
)

type QueryHandler interface {
	ExecuteQuery(user.User, string) (interface{}, error)
}

type postgresQueryHandler struct {
	db *sql.DB
}

func NewPostgresQueryHandler() QueryHandler {
	return &postgresQueryHandler{}
}

func (h *postgresQueryHandler) ExecuteQuery(usr user.User, sql string) (interface{}, error) {
	res, err := h.db.Query(sql)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
