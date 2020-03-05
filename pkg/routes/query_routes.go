package routes

import (
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/griffin/cs622-datasec/pkg/audit"
	"github.com/griffin/cs622-datasec/pkg/user"
)

type QueryManager struct {
	//Policy PolicyHandler
	Audit audit.AuditHandler
	//Verify VerifyHandler
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

	err = qm.Audit.LogQuery(user.User{}, audit.QueryRecieved, query.SQL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{
			Message: "could not save audit",
		})
		log.Warn(err)
		return
	}

}
