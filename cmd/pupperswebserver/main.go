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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/natemarks/puppers"
	"github.com/natemarks/puppers/secrets"

	_ "github.com/lib/pq"
	"github.com/natemarks/postgr8/command"

	"github.com/rs/zerolog"
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

func heartbeat(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "Status OK"
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}

func wait(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	waitDuration := r.URL.Query().Get("wait")
	wait, err := time.ParseDuration(waitDuration)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		resp["message"] = "Invalid wait parameter example 500ms"
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}
	time.Sleep(wait)
	w.WriteHeader(http.StatusOK)
	resp["message"] = fmt.Sprintf("You waited for %s", wait)
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)

	return
}

func main() {
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			log.Panic().Msg(err.Error())
		}
	}(logFile)
	log.Info().Msgf("pupperswebserver is starting with graceful shutdown timeout: %s",
		gracefulShutdownTimeout)
	mux := http.NewServeMux()

	mux.HandleFunc("/", wait)
	mux.HandleFunc("/heartbeat", heartbeat)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Panic().Msgf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Panic().Msgf("Server Shutdown Failed: %s", err.Error())
	}
	log.Info().Msg("Graceful shutdown complete")
}
