package main

// Package main is a test webserver with several useful capabilities:
// - it captures SIGINT /SIGTERM signals to start a graceful shutdown
//
// - graceful shutdown log messages can be used to confirm that all transactions
// were completed before shutting down
//
// - while running it serves a mock API that only accepts a wait time parameter.
// this keeps the transaction running for a known duration, which is useful for
// testing the success/failure of graceful.
//
// NOTE: ECS kills tasks 30 seconds after sending the polite SIGINT. The default
// internal timeout is 200s. it can be overridden with GRACEFUL_SHUTDOWN_TIMEOUT
//
// shutdown - health/ready check
import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/natemarks/puppers"
	"github.com/natemarks/puppers/secrets"

	_ "github.com/lib/pq"
	"github.com/natemarks/postgr8/command"

	"github.com/rs/zerolog"

	ginzerolog "github.com/dn365/gin-zerolog"
)

const defaultGracefulShutdownTimeout = "200s"

var gracefulShutdownTimeout time.Duration
var err error
var creds command.InstanceConnectionParams
var logFile *os.File
var log zerolog.Logger

// Check all dependencies at startup
func init() {
	logFile, err = os.OpenFile("pupperswebserver.log",
		os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	// configure logger
	log = zerolog.New(logFile).With().Timestamp().Logger()
	log = log.With().Str("version", puppers.Version).Logger()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// set gracefulShutdownTimeout
	gracefulShutdownTimeout, err = time.ParseDuration(os.Getenv("GRACEFUL_SHUTDOWN_TIMEOUT"))
	if err != nil {
		gracefulShutdownTimeout, _ = time.ParseDuration(defaultGracefulShutdownTimeout)
	}

	// make sure we can access the credentials
	creds = secrets.GetRDSCredentials(secrets.GetSecretFromEnvar(), &log)

	// Check connectivity to database instance
	if !command.TCPOk(creds, 5) {
		log.Panic().Msgf("TCP Connection Failure: %s:%d", creds.Host, creds.Port)
	}
	log.Info().Msgf("TCP Connection Success: %s:%d", creds.Host, creds.Port)

	// Make sure the credentials are valid
	validCreds, err := command.ValidCredentials(creds)
	if !validCreds {
		log.Panic().Msg(err.Error())
	}
	log.Info().Msg("database credentials are valid")
	// run test query to check the number of databases
	conn, err := command.NewInstanceConn(creds)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	dbNames, err := command.ListDatabases(conn)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	log.Info().Msgf("Found databases in instance: %d", len(dbNames))
}

func waitResponse(w string) string {
	if wait, err := time.ParseDuration(w); err == nil {
		time.Sleep(wait)
		return fmt.Sprintf("You waited for %s", w)
	}
	return "Invalid wait parameter example 500ms"
}

func main() {
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {

		}
	}(logFile)
	log.Info().Msgf("pupperswebserver is starting with graceful shutdown timeout: %s",
		gracefulShutdownTimeout)

	router := gin.Default()
	router.Use(ginzerolog.Logger("gin"), gin.Recovery())
	router.GET("/", func(c *gin.Context) {
		q := c.Request.URL.Query()
		c.String(http.StatusOK, waitResponse(fmt.Sprint(q["wait"][0])))
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	hbRouter := gin.Default()
	hbRouter.Use(ginzerolog.Logger("gin"), gin.Recovery())

	hbRouter.GET("/heartbeat", func(c *gin.Context) {
		c.Data(200, "text/plain", []byte("."))
	})

	hbSrv := &http.Server{
		Addr:    ":8786",
		Handler: hbRouter,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("listen: %s\n", err)
		}
	}()

	go func() {
		// service connections
		if err := hbSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Msgf("Server Shutdown: %s", err.Error())
	}
	if err := hbSrv.Shutdown(ctx); err != nil {
		log.Fatal().Msgf("Server Shutdown: %s", err.Error())
	}

	log.Info().Msg("Graceful shutdown complete")
}
