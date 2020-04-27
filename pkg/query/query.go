package query

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/xwb1989/sqlparser"

	"github.com/griffin/cs622-datasec/pkg/user"
)

var (
	ErrIncorrectType = fmt.Errorf("incorrect statement type")
	ErrUnknownRole   = fmt.Errorf("unknown role")
)

type Val struct {
	Column string      `json:"column"`
	Value  interface{} `json:"value"`
}

type QueryHandler interface {
	Query(user.User, string) ([][]*Val, error)
	Exec(user.User, string) error
}

type postgresQueryHandler struct {
	db *sql.DB
}

func NewPostgresQueryHandler(db *sql.DB) QueryHandler {
	return &postgresQueryHandler{
		db: db,
	}
}

func makeColummnsArray(cols []string) ([]*Val, []interface{}) {
	var rtVal []*Val
	var rtInf []interface{}

	for _, c := range cols {
		v := &Val{
			Column: c,
		}
		rtVal = append(rtVal, v)
		rtInf = append(rtInf, &v.Value)
	}

	return rtVal, rtInf
}

func getStatementType(sql string) int {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		log.Println(err)
		return sqlparser.StmtOther
	}

	switch stmt.(type) {
	case *sqlparser.Select:
		return sqlparser.StmtSelect

	case *sqlparser.Insert:
		return sqlparser.StmtInsert
	}

	return sqlparser.StmtOther
}

func (h *postgresQueryHandler) Query(usr user.User, sql string) ([][]*Val, error) {
	ty := getStatementType(sql)
	if ty != sqlparser.StmtSelect {
		return nil, ErrIncorrectType
	}

	tx, err := h.db.Begin()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(fmt.Sprintf("SET ROLE %v;", usr.PostgresUser))
	if err != nil {
		return nil, ErrUnknownRole
	}

	rows, err := tx.Query(sql)
	if err != nil {
		return nil, err
	}

	colNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var rt [][]*Val

	for rows.Next() {
		r, scans := makeColummnsArray(colNames)

		if err := rows.Scan(scans...); err != nil {
			return nil, err
		}

		rt = append(rt, r)
	}

	_, err = tx.Exec("RESET ROLE;")
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return rt, nil
}

func (h *postgresQueryHandler) Exec(usr user.User, sql string) error {
	ty := getStatementType(sql)
	if ty != sqlparser.StmtInsert {
		return ErrIncorrectType
	}

	tx, err := h.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("SET ROLE $1;", usr.PostgresUser)
	if err != nil {
		return ErrUnknownRole
	}

	_, err = tx.Exec(sql)
	if err != nil {
		return err
	}

	_, err = tx.Exec("RESET ROLE;")
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
