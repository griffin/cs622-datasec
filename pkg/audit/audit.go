package audit

import (
	"fmt"
	"net/http"
	"net/url"

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

type httpAuditHandler struct {
	baseURL *url.URL
}

func NewHttpAuditHandler(url *url.URL) AuditHandler {
	return &httpAuditHandler{
		baseURL: url,
	}
}

func (h *httpAuditHandler) LogQuery(usr user.User, status QueryStatus, query string) error {
	resp, err := http.PostForm(
		fmt.Sprintf("%v/audit", h.baseURL.Host),
		url.Values{"sql": []string{query}, "user": []string{usr.PostgreUser}, "status": []string{string(status)}},
	)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to log query")
	}

	return nil
}
