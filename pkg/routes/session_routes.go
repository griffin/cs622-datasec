package routes

import (
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/griffin/cs622-datasec/pkg/session"
	"github.com/griffin/cs622-datasec/pkg/user"
	"net/http"
)

type UserManager struct {
	User    user.UserDatastore
	Session session.SessionDatastore
}

type UserCreate struct {
	Name     string
	Email    string
	Password string
}

type Error struct {
	Message string
}

type TokenResponse struct {
	Token string
}

func (uc UserCreate) ToUser() user.User {
	return user.User{
		Name:  uc.Name,
		Email: uc.Email,
	}
}

type UserLogin struct {
	Email    string
	Password string
}

func (um *UserManager) PostRegisterRoute(c *gin.Context) {
	userCreate := UserCreate{}
	err := c.Bind(&userCreate)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{
			"err": "couldn't create user",
		})
		log.Warn(err)
		return
	}

	_, err = um.User.Create(userCreate.ToUser(), userCreate.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{
			"err": "couldn't create user",
		})
		log.Warn(err)
		return
	}

	c.Status(http.StatusOK)
}

func (um *UserManager) PostLoginRoute(c *gin.Context) {
	userLogin := UserLogin{}
	err := c.Bind(&userLogin)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{
			Message: "incorrect json format",
		})
		log.Warn(err)
		return
	}

	token, err := um.Session.Create(userLogin.Email, userLogin.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Error{
			Message: "could not authenticate",
		})
		log.Warn(err)
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		Token: token,
	})
}

func (um *UserManager) PostLogoutRoute(c *gin.Context) {
	// TODO(griffin): logout not a priority
}
