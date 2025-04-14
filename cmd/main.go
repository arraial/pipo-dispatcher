package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	common "github.com/arraial/pipo-dispatcher/internal/common"
	"github.com/arraial/pipo-dispatcher/internal/queues"
	"github.com/arraial/pipo-dispatcher/models"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var version string = "latest"
var connections []*amqp.ConnectionWrapper

func main() {
	if err := run(); err != nil {
		common.GetLogger().Error(err.Error())
	}
}

func serverIsHealthy() bool {
	return true
}

func livezHandler(w http.ResponseWriter, r *http.Request) {
	if serverIsHealthy() {
		w.WriteHeader(http.StatusOK)
		common.GetLogger().Info("Server is healthy")
	}
}

func serverIsReady() bool {
	for _, conn := range connections {
		if conn == nil || !conn.IsConnected() {
			return false
		}
	}
	return true
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

	otelShutdown, err := common.SetupOTelSDK(ctx)
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
	messages := make(chan *models.ProviderOperation, config.GetInt("queue.broker.buffer_size"))
	defer close(messages)

	consumer, err := queues.CreateConsumer(ctx, messages)
	if err != nil {
		log.Fatalw("Failed to start consumer", "error", err)
	}

	publisher, err := queues.CreatePublisher(ctx, messages)
	if err != nil {
		log.Fatalw("Failed to start publisher", "error", err)
	}

	connections = append(connections, consumer.ConnectionWrapper, publisher.ConnectionWrapper)

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
