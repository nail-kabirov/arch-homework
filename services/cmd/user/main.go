package main

import (
	"arch-homework/pkg/common/infrastructure/metrics"
	commonpostgres "arch-homework/pkg/common/infrastructure/postgres"
	"arch-homework/pkg/common/jwtauth"
	"arch-homework/pkg/user/app"
	"arch-homework/pkg/user/infrastructure/postgres"
	serverhttp "arch-homework/pkg/user/infrastructure/transport/http"

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

	metricsHandler, err := metrics.NewPrometheusMetricsHandler(serverhttp.NewEndpointLabelCollector())
	if err != nil {
		logger.Fatal(err)
	}

	server := startServer(cfg, connector, logger, metricsHandler)

	waitForKillSignal(logger)
	if err := server.Shutdown(context.Background()); err != nil {
		logger.WithError(err).Fatal("http server shutdown failed")
	}
}

func initDBConnector(cfg *config) (commonpostgres.Connector, error) {
	connector := commonpostgres.NewConnector()
	err := connector.Open(commonpostgres.DSN{
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

func startServer(cfg *config, connector commonpostgres.Connector, logger *logrus.Logger, metricsHandler metrics.PrometheusMetricsHandler) *http.Server {
	httpAddress := ":" + cfg.ServicePort
	if err := connector.WaitUntilReady(); err != nil {
		logger.Fatal(err)
	}
	dbDependency := postgres.NewDBDependency(connector.Client())
	userService := app.NewUserService(dbDependency)
	tokenParser := jwtauth.NewTokenParser(cfg.JWTSecret)
	userServer := serverhttp.NewServer(userService, tokenParser, logger)

	router := mux.NewRouter()
	router.HandleFunc("/health", handleHealth).Methods(http.MethodGet)
	router.HandleFunc("/ready", handleReady(connector)).Methods(http.MethodGet)
	router.PathPrefix(serverhttp.PathPrefix).Handler(userServer.MakeHandler())

	metricsHandler.AddMetricsHandler(router, "/metrics")
	metricsHandler.AddCommonMetricsMiddleware(router)

	server := &http.Server{
		Handler:      router,
		Addr:         httpAddress,
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
	}

	go func() {
		logger.Fatal(server.ListenAndServe())
	}()

	return server
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, "{\"status\": \"OK\"}")
}

func handleReady(connector commonpostgres.Connector) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if connector.Ready() {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, http.StatusText(http.StatusOK))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}
}
