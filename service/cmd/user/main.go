package main

import (
	"arch-homework1/pkg/user/app"
	"arch-homework1/pkg/user/infrastructure/postgres"
	serverhttp "arch-homework1/pkg/user/infrastructure/transport/http"

	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	ReadTimeout  = time.Minute
	WriteTimeout = time.Minute
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.Info("service started")

	cfg, err := parseEnv()
	if err != nil {
		logger.Fatal(err)
	}

	connector, err := initDBConnector(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer connector.Close()

	server := startServer(":"+cfg.ServicePort, connector, logger)

	waitForKillSignal(logger)
	if err := server.Shutdown(context.Background()); err != nil {
		logger.WithError(err).Fatal("http server shutdown failed")
	}
}

func initDBConnector(cfg *config) (postgres.Connector, error) {
	connector := postgres.NewConnector()
	err := connector.Open(postgres.DSN{
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		Database: cfg.DBName,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open database")
	}
	return connector, err
}

func waitForKillSignal(logger *logrus.Logger) {
	sysKillSignal := make(chan os.Signal, 1)
	signal.Notify(sysKillSignal, os.Interrupt, syscall.SIGTERM)
	logger.Infof("got system signal '%s'", <-sysKillSignal)
}

func startServer(httpAddress string, connector postgres.Connector, logger *logrus.Logger) *http.Server {
	router := mux.NewRouter()
	router.HandleFunc("/health", handleHealth).Methods(http.MethodGet)
	router.HandleFunc("/ready", handleReady(connector)).Methods(http.MethodGet)
	serveMux := http.NewServeMux()
	serveMux.Handle("/", router)

	server := &http.Server{
		Handler:      serveMux,
		Addr:         httpAddress,
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
	}

	go func() {
		logger.Fatal(server.ListenAndServe())
	}()

	go func() {
		if err := connector.WaitUntilReady(); err != nil {
			logger.Fatal(err)
		}
		userService := app.NewUserService(postgres.NewUSerRepository(connector.Client()))
		server := serverhttp.NewServer(userService, logger)
		serveMux.Handle(serverhttp.PathPrefix, server.MakeHandler())
	}()

	return server
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, "{\"status\": \"OK\"}")
}

func handleReady(connector postgres.Connector) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if connector.Ready() {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, http.StatusText(http.StatusOK))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}
}
