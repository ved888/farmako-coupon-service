package main

import (
	"farmako-coupon-service/cache"
	"farmako-coupon-service/database"
	"farmako-coupon-service/docs"
	"farmako-coupon-service/server"
	"farmako-coupon-service/utils"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/sirupsen/logrus"
)

const (
	shutDownTimeOut = 10 * time.Second
)

func init() {
	rand.Seed(time.Now().UnixNano())
	cache.Init()

}

// @title           farmako-coupon-service
// @version         1.0
// @description     This is the main server handling the farmako-coupon-service major operations.
// @contact.name   farmako-coupon-service
// @contact.url    https://farmako-coupon-service.com/
// @in header
// @name x-api-key
func main() {
	// setup logger
	logrus.SetFormatter(&logrus.JSONFormatter{PrettyPrint: !utils.IsBranchEnvSet()}) // only pretty print on local
	logrus.SetReportCaller(true)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// create server instance
	srv := server.SetupBaseV1Routes()

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	if err := database.ConnectAndMigrate(os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		database.SSLModeDisable); err != nil {
		logrus.WithError(err).Panic("Failed to initialize and migrate database")
	}
	logrus.Info("database connection and migration successful...")

	go func() {
		// setup swagger route only on dev or local development
		if !utils.IsBranchEnvSet() || utils.GetBranch() == utils.Development {
			if !utils.IsBranchEnvSet() {
				docs.SwaggerInfo.Schemes = []string{"http"}
				docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", "localhost", os.Getenv("PORT"))
				docs.SwaggerInfo.BasePath = "/"
			} else {
				docs.SwaggerInfo.Schemes = []string{"https"}
				docs.SwaggerInfo.Host = "dev-api.fcs.com"
				docs.SwaggerInfo.BasePath = "/"
			}
			srv.Router.Route("/docs", func(r chi.Router) {
				r.Get("/*", httpSwagger.Handler())
			})

		}

		// start the server
		if err := srv.Run(":" + os.Getenv("PORT")); err != nil && err != http.ErrServerClosed {
			logrus.Panicf("Failed to run server with error: %+v", err)
		}
	}()

	logrus.Printf("Running on prod: %+v", utils.IsProd())
	logrus.Print("Server started at ", os.Getenv("PORT"))

	<-done

	logrus.Info("shutting down server")

	if err := database.ShutdownDatabase(); err != nil {
		logrus.WithError(err).Error("failed to close database connection")
	}

	if err := srv.Shutdown(shutDownTimeOut); err != nil {
		logrus.WithError(err).Panic("failed to gracefully shutdown server")
	}

}
