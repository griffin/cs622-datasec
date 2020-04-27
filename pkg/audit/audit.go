package audit

import (
	"database/sql"
	"fmt"

	"github.com/griffin/cs622-datasec/pkg/user"
)

type QueryStatus string

const (
	QuerySuccess             QueryStatus = "success"
	QueryFailedPolicy        QueryStatus = "failed_policy"
	QueryFailedAuthorization QueryStatus = "failed_authorization"
)

const (
	insertLog = "INSERT INTO audit_query (user_id, postgres_user, status, query) VALUES ($1, $2, $3, $4)"
)

type AuditHandler interface {

	// LogQuery logs any query sent to  the database for inspection
	LogQuery(usr user.User, status QueryStatus, query string) error
}

type dbAuditHandler struct {
	sqlClient *sql.DB
}

func NewDBAuditHandler(db *sql.DB) AuditHandler {
	return &dbAuditHandler{
		sqlClient: db,
	}
}

func (h *dbAuditHandler) LogQuery(usr user.User, status QueryStatus, query string) error {
	_, err := h.sqlClient.Exec(insertLog, usr.ID, usr.PostgresUser, status, query)
	if err != nil {
		return fmt.Errorf("Could not insert log: %v", err)
	}

	return nil
}
