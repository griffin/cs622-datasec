package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/griffin/cs622-datasec/pkg/audit"
	"github.com/griffin/cs622-datasec/pkg/datastore"
	"github.com/griffin/cs622-datasec/pkg/policy"
	"github.com/griffin/cs622-datasec/pkg/query"
	"github.com/griffin/cs622-datasec/pkg/routes"
	"github.com/griffin/cs622-datasec/pkg/session"
	"github.com/griffin/cs622-datasec/pkg/user"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

type Options struct {
	SQLUsername string `long:"sql_username" description:"sql account username"`
	SQLPassword string `long:"sql_password" description:"sql account password"`
	SQLDatabase string `long:"sql_database" description:"sql database"`
	SQLHost     string `long:"sql_host" description:"hostname of sql database"`
	SQLPort     uint   `long:"sql_port" default:"5432" description:"sql port"`
	Port        uint   `long:"port" default:"8080" description:"api port"`
	PolicyFile  string `long:"policy_file"`
}

func (opts Options) GetSQLOptions() datastore.SQLOptions {
	return datastore.SQLOptions{
		User:     opts.SQLUsername,
		Password: opts.SQLPassword,
		Host:     opts.SQLHost,
		Port:     opts.SQLPort,
		Database: opts.SQLDatabase,
	}
}

func main() {
	var opts Options
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Fatal(err)
	}

	ds, err := datastore.New(opts.GetSQLOptions())
	if err != nil {
		log.Fatalf("failed to init datastore: %v", err)
	}

	router := gin.Default()
	router.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))

	a := audit.NewDBAuditHandler(ds.SQLClient)

	content, err := ioutil.ReadFile(opts.PolicyFile)
	if err != nil {
		log.Fatal(err)
	}

	p := policy.NewHttpPolicyHandler(string(content))

	s := session.NewSessionDatastoreHandler(ds.SQLClient)
	um := routes.UserManager{
		User:    user.NewUserDatastoreHandler(ds.SQLClient),
		Session: s,
	}

	q := query.NewPostgresQueryHandler(ds.SQLClient)

	qm := routes.QueryManager{
		Policy: p,
		Audit:  a,
		Query:  q,
	}

	router.POST("/v1/login", um.PostLoginRoute)
	router.POST("/v1/logout", um.PostLogoutRoute)
	router.POST("/v1/register", um.PostRegisterRoute)

	router.POST("/v1/query", authMiddeware(qm.PostQueryRoute, s))

	log.Fatal(router.Run(fmt.Sprintf(":%v", opts.Port)))
}

func authMiddeware(fn func(*gin.Context), s session.SessionDatastore) func(c *gin.Context) {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			log.Warn("no auth given")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		split := strings.Split(auth, " ")

		if len(split) < 2 {
			log.Warn("not in correct format")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if split[0] != "Bearer" {
			log.Warn("not a Bearer token")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		usr, err := s.Check(split[1])
		if err != nil {
			log.Warn(err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		fmt.Println(usr)

		c.Set("user", usr)

		fn(c)
	}
}

func injectUserMiddleware(fn func(*gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Set("user", user.User{
			Name:         "test.user",
			Email:        "test.user@example.com",
			PostgresUser: "postgres",
		})

		fn(c)
	}
}
