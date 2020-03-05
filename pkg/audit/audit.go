package audit

import (
	"github.com/griffin/cs622-datasec/pkg/user"
)

type QueryStatus string

const (
	QueryRecieved            QueryStatus = "recieved"
	QuerySuccess             QueryStatus = "success"
	QueryFailedPolicy        QueryStatus = "failed_policy"
	QueryFailedValidation    QueryStatus = "failed_validation"
	QueryFailedAudit         QueryStatus = "failed_audit"
	QueryFailedAuthorization QueryStatus = "failed_authorization"
)

type AuditHandler interface {

	// LogQuery logs any query sent to  the database for inspection
	LogQuery(usr user.User, status QueryStatus, query string) error
}
