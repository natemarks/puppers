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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
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

}

func waitResponse(w string) string {
	if wait, err := time.ParseDuration(w); err == nil {
		time.Sleep(wait)
		return fmt.Sprintf("You waited for %s", w)
	}
	return "Invalid wait parameter example 500ms"
}

func hearbeat(w http.ResponseWriter, r *http.Request) {
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

	http.HandleFunc("/", wait)
	http.HandleFunc("/heartbeat", hearbeat)

	err := http.ListenAndServe(":8080", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
