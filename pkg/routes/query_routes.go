package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/griffin/cs622-datasec/pkg/audit"
	"github.com/griffin/cs622-datasec/pkg/policy"
	"github.com/griffin/cs622-datasec/pkg/query"
	"github.com/griffin/cs622-datasec/pkg/user"
)

type QueryManager struct {
	Policy policy.PolicyHandler
	Audit  audit.AuditHandler
	Query  query.QueryHandler
}

type Query struct {
	SQL string `json:"sql"`
}

func (qm *QueryManager) PostQueryRoute(c *gin.Context) {
	query := Query{}
	err := c.Bind(&query)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{
			Message: "incorrect json format",
		})
		log.Warn(err)
		return
	}

	if query.SQL == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{
			Message: "no query given",
		})
		log.Warn("no query given")
		return

	}

	usrVal, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{
			Message: "failed to find user",
		})
		log.Warn("couldn't find user in context")
		return
	}

	usr, ok := usrVal.(user.User)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Error{
			Message: "failed to find user",
		})
		log.Warnf("failed to find user %v", usr)
		return
	}

	var status audit.QueryStatus
	defer func(s *audit.QueryStatus) {
		err = qm.Audit.LogQuery(usr, *s, query.SQL)
		if err != nil {
			log.Warn(err)
		}
	}(&status)

	err = qm.Policy.CheckPolicy(usr, query.SQL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{
			Message: "failed policy check",
		})
		log.Warn(err)

		status = audit.QueryFailedPolicy
		return
	}

	res, err := qm.Query.Query(usr, query.SQL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{
			Message: "failed query",
		})
		log.Warn(err)

		status = audit.QueryFailedAuthorization
		return
	}

	status = audit.QuerySuccess

	c.AbortWithStatusJSON(http.StatusOK, res)
}
