package main

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-resty/resty/v2"
	"github.com/hack-fan/config"
	"github.com/hack-fan/x/xdb"
	"github.com/hyacinthus/x/xerr"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/hack-fan/skadi/service"
	"github.com/hack-fan/skadi/types"
)

func main() {
	var err error
	// load config
	var settings = new(Settings)
	config.MustLoad(settings)

	// logger
	var logger *zap.Logger
	if settings.Debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // nolint
	var log = logger.Sugar()

	// kv
	var kv = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", settings.Redis.Host, settings.Redis.Port),
		Password: settings.Redis.Password,
		DB:       settings.Redis.DB,
	})

	// db
	xdb.SetLogger(log)
	var db = xdb.New(settings.DB)
	if settings.Debug {
		db = db.Debug()
	}
	// auto create table
	go db.AutoMigrate(&types.Job{}) // nolint

	// http client
	var rest = resty.New().SetRetryCount(3).
		SetRetryWaitTime(5 * time.Second).
		SetRetryMaxWaitTime(60 * time.Second)

	// job service
	var js = service.New(kv, db, rest, log)

	// handler
	var h = NewHandler(js)

	// Echo instance
	e := echo.New()
	if settings.Debug {
		e.Debug = true
	}
	// Error handler
	e.HTTPErrorHandler = xerr.ErrorHandler
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.KeyAuth(js.AuthValidator))

	// Routes
	e.GET("/", getStatus)

	e.GET("/agent/job", h.GetJob)
	e.PUT("/agent/jobs/:id/succeed", h.PutJobSucceed)
	e.PUT("/agent/jobs/:id/fail", h.PutJobFail)

	// Start server
	e.Logger.Fatal(e.Start(settings.ListenAddr))

}
