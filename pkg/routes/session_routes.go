package routes

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/griffin/cs622-datasec/pkg/session"
	"github.com/griffin/cs622-datasec/pkg/user"
)

type UserManager struct {
	User    user.UserDatastore
	Session session.SessionDatastore
}

type UserCreate struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Error struct {
	Message string `json:"message"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

func (uc UserCreate) ToUser() user.User {
	return user.User{
		Name:  uc.Name,
		Email: uc.Email,
	}
}

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

	token, err := um.Session.Create(userLogin.Email, userLogin.Password, time.Hour*2)
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
