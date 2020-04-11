package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/griffin/cs622-datasec/pkg/audit"
	"github.com/griffin/cs622-datasec/pkg/policy"
	"github.com/griffin/cs622-datasec/pkg/query"
	"github.com/griffin/cs622-datasec/pkg/user"
	//	"github.com/griffin/cs622-datasec/pkg/verify"
)

type QueryManager struct {
	Policy policy.PolicyHandler
	Audit  audit.AuditHandler
	//	Verify verify.VerifyHandler
	Query query.QueryHandler
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

	err = qm.Audit.LogQuery(usr, audit.QueryRecieved, query.SQL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{
			Message: "could not save audit",
		})
		log.Warn(err)
		return
	}

	err = qm.Policy.CheckPolicy(usr, query.SQL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{
			Message: "failed policy check",
		})
		log.Warn(err)
		return
	}

	res, err := qm.Query.ExecuteQuery(usr, query.SQL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{
			Message: "failed query",
		})
		log.Warn(err)
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, res)
}
