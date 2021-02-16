// Copyright (c) J.Dreyer
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/jonathandreyer/mainflux-httpforwarder/http-forwarder"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/messaging/nats"
	"github.com/mainflux/mainflux/transformers/senml"
	"github.com/mainflux/mainflux/writers"
	"github.com/mainflux/mainflux/writers/api"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	svcName = "http-forwarder"

	defNatsURL         = "nats://localhost:4222"
	defLogLevel        = "error"
	defPort            = "8990"
	defRemoteUrl       = "http://localhost:9000"
	defRemoteToken     = ""
	defSubjectsCfgPath = "/config/subjects.toml"
	defContentType     = "application/senml+json"

	envNatsURL         = "MF_NATS_URL"
	envLogLevel        = "MF_HTTP_FORWARDER_LOG_LEVEL"
	envPort            = "MF_HTTP_FORWARDER_PORT"
	envRemoteUrl       = "MF_HTTP_FORWARDER_REMOTE_URL"
	envRemoteToken     = "MF_HTTP_FORWARDER_REMOTE_TOKEN"
	envSubjectsCfgPath = "MF_HTTP_FORWARDER_SUBJECTS_CONFIG"
	envContentType     = "MF_HTTP_FORWARDER_CONTENT_TYPE"
)

type config struct {
	natsURL         string
	logLevel        string
	port            string
	remoteUrl       string
	remoteToken     string
	subjectsCfgPath string
	contentType     string
}

func main() {
	cfg := loadConfigs()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	pubSub, err := nats.NewPubSub(cfg.natsURL, "", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer pubSub.Close()

	repo := http_forwarder.New(cfg.remoteUrl, cfg.remoteToken)

	counter, latency := makeMetrics()
	repo = api.LoggingMiddleware(repo, logger)
	repo = api.MetricsMiddleware(repo, counter, latency)
	st := senml.New(cfg.contentType)
	if err := writers.Start(pubSub, repo, st, svcName, cfg.subjectsCfgPath, logger); err != nil {
		logger.Error(fmt.Sprintf("Failed to start HTTP forwarder: %s", err))
		os.Exit(1)
	}

	errs := make(chan error, 2)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go startHTTPService(cfg.port, logger, errs)

	err = <-errs
	logger.Error(fmt.Sprintf("HTTP forwarder service terminated: %s", err))
}

func loadConfigs() config {
	cfg := config{
		natsURL:         mainflux.Env(envNatsURL, defNatsURL),
		logLevel:        mainflux.Env(envLogLevel, defLogLevel),
		port:            mainflux.Env(envPort, defPort),
		remoteUrl:       mainflux.Env(envRemoteUrl, defRemoteUrl),
		remoteToken:     mainflux.Env(envRemoteToken, defRemoteToken),
		subjectsCfgPath: mainflux.Env(envSubjectsCfgPath, defSubjectsCfgPath),
		contentType:     mainflux.Env(envContentType, defContentType),
	}

	return cfg
}

func makeMetrics() (*kitprometheus.Counter, *kitprometheus.Summary) {
	counter := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "http_forwarder",
		Subsystem: "message_writer",
		Name:      "request_count",
		Help:      "Number of database inserts.",
	}, []string{"method"})

	latency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "http_forwarder",
		Subsystem: "message_writer",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of inserts in microseconds.",
	}, []string{"method"})

	return counter, latency
}

func startHTTPService(port string, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", port)
	logger.Info(fmt.Sprintf("HTTP forwarder service started, exposed port %s", p))
	errs <- http.ListenAndServe(p, api.MakeHandler(svcName))
}
