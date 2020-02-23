package main

import (
	log "github.com/sirupsen/logrus"

	"fmt"
	"os"
	"time"

	"github.com/griffin/cs622-datasec/pkg/datastore"
	"github.com/griffin/cs622-datasec/pkg/routes"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	SQLUsername string
	SQLPassword string
	SQLDatabase string
	SQLHost     string
	SQLPort     uint
	Port        uint
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

	um := routes.UserManager{
		User:    ds.User,
		Session: ds.Session,
	}

	router.POST("/v1/login", um.PostLoginRoute)
	log.Fatal(router.Run(fmt.Sprintf(":%v", opts.Port)))
}
