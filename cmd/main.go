package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	amqpMessage "github.com/ThreeDotsLabs/watermill/message"
	common "github.com/arraial/pipo-dispatcher/internal/common"
	"github.com/arraial/pipo-dispatcher/internal/queues"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var version string = "latest"
var router *amqpMessage.Router

func main() {
	if err := run(); err != nil {
		common.GetLogger().Error(err.Error())
	}
}

func serverIsHealthy() bool {
	return router != nil
}

func livezHandler(w http.ResponseWriter, r *http.Request) {
	if serverIsHealthy() {
		w.WriteHeader(http.StatusOK)
		common.GetLogger().Info("Server is healthy")
	}
}

func serverIsReady() bool {
	return router != nil && !router.IsClosed() && router.IsRunning()
}

func readyzHandler(w http.ResponseWriter, r *http.Request) {
	if serverIsReady() {
		w.WriteHeader(http.StatusOK)
		common.GetLogger().Info("Server is ready")
	}
}

func newHTTPHandler() http.Handler {
	config := common.GetConfig()

	mux := http.NewServeMux()

	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	handleFunc(config.GetString("probes.liveness.endpoint"), livezHandler)
	handleFunc(config.GetString("probes.readiness.endpoint"), readyzHandler)

	return otelhttp.NewHandler(mux, config.GetString("telemetry.metrics.endpoint"))
}

func run() (err error) {
	config := common.GetConfig()

	log := common.GetLogger()
	defer log.Sync()

	log.Infow("Starting application", "app", config.GetString("app"), "version", version)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	errs := make(chan error, 1)

	otelShutdown, err := common.SetupOTelSDK(ctx, config.GetString("telemetry.service"), version)
	if err != nil {
		log.Errorw("Failed to initialize OTel SDK", err)
		return
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
		log.Error("Unable to shutdown otel")
	}()

	srv := &http.Server{
		Addr:         config.GetString("probes.host") + ":" + config.GetString("probes.port"),
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  config.GetDuration("probes.timeout.read"),
		WriteTimeout: config.GetDuration("probes.timeout.write"),
		Handler:      newHTTPHandler(),
	}
	go func() {
		err = errors.Join(err, srv.ListenAndServe())
	}()
	defer func() {
		log.Errorw("Unable to shutdown probe server", "error", srv.Shutdown(ctx))
	}()

	log.Info("Probe HTTP server started")

	router, err = queues.CreateRouter(ctx)
	if err != nil {
		log.Fatalw("Failed to create router", "error", err)
	}

	go func() {
		err = errors.Join(err, router.Run(ctx))
	}()

	log.Info("Started processing messages")
	for {
		select {
		case err = <-errs:
			log.Errorw("Unexpected error", "error", err)
			stop()
		case <-ctx.Done():
			err = errors.Join(err, ctx.Err())
			log.Infow("Stopping application", "error", err)
			return
		}
	}
}
