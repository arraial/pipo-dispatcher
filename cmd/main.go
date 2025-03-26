package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	common "github.com/arraial/pipo-dispatcher/internal/common"
	queues "github.com/arraial/pipo-dispatcher/internal/queues"
	"github.com/arraial/pipo-dispatcher/models"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

var version string = "latest"

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

// TODO check connection to MQ
func readyzHandler(w http.ResponseWriter, r *http.Request) {
	if serverIsHealthy() {
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

	logger, _ := zap.NewProduction()
	defer logger.Sync()
	log := logger.Sugar()

	log.Infow("Starting application", "app", config.GetString("app"), "version", version)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	errs := make(chan error, 5)
	var rt_err error

	otelShutdown, err := common.SetupOTelSDK(ctx)
	if err != nil {
		log.Error("Failed to initialize OTel SDK", err)
		errs <- err
	}
	defer func() {
		rt_err = errors.Join(rt_err, otelShutdown(context.Background()))
	}()

	srv := &http.Server{
		Addr:         config.GetString("probes.host") + ":" + config.GetString("probes.port"),
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  config.GetDuration("probes.timeout.read"),
		WriteTimeout: config.GetDuration("probes.timeout.write"),
		Handler:      newHTTPHandler(),
	}
	go func() {
		errs <- srv.ListenAndServe()
	}()
	defer func() {
		rt_err = errors.Join(rt_err, srv.Shutdown(ctx))
	}()

	log.Info("Probe HTTP server started")
	messages := make(chan *models.ProviderOperation, config.GetInt("queue.broker.buffer_size"))
	defer close(messages)

	connection, err := queues.Connection(config.GetString("queue.broker.url"))
	if err != nil {
		log.Errorw("Failed to initialize message queue connection", "error", err, "url")
		errs <- err
	} else {
		defer connection.Close()
	}

	publisher, err := queues.StartPublisher(ctx, connection, messages)
	if err != nil {
		log.Errorw("Failed to start publisher", "error", err)
		errs <- err
	} else {
		defer publisher.Close()
	}

	consumer, err := queues.StartConsumer(ctx, connection, messages, config.GetInt("queue.broker.max_consumers"))
	if err != nil {
		log.Errorw("Failed to start consumer", "error", err)
		errs <- err
	} else {
		defer consumer.Close()
	}

	log.Info("Started processing messages")
	for {
		select {
		case err = <-errs:
			log.Errorw("Unexpected error", "error", err)
			stop()
		case <-ctx.Done():
			rt_err = ctx.Err()
			log.Infow("Stopping application", "error", err)
			return rt_err
		}
	}
}
