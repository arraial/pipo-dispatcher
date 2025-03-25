package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/arraial/pipo-dispatcher/internal/settings"
	"github.com/arraial/pipo-dispatcher/internal/telemetry"
	config "github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var version string = "latest"

func main() {
	var conf = settings.InitConfig()
	log.Println(conf.GetString("telemetry.service"))
	log.Println("Version: " + version)

	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func serverIsHealthy() bool {
	return true
}

func livezHandler(w http.ResponseWriter, r *http.Request) {
	if serverIsHealthy() {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Server is healthy")
	}
}

func readyzHandler(w http.ResponseWriter, r *http.Request) {
	if serverIsHealthy() {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Server is healthy")
	}
}

func newHTTPHandler() http.Handler {
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	otelShutdown, err := telemetry.SetupOTelSDK(ctx)
	if err != nil {
		return
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// Start HTTP server
	srv := &http.Server{
		Addr:         config.GetString("probes.host") + ":" + config.GetString("probes.port"),
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  config.GetDuration("probes.timeout.read"),
		WriteTimeout: config.GetDuration("probes.timeout.write"),
		Handler:      newHTTPHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// TODO Handle interruptions
	select {
	case err = <-srvErr:
		log.Fatal("Unexpected error happened. Stopping application.")
		return
	case <-ctx.Done():
		log.Println("Application was stopped.")
		stop()
	}

	err = srv.Shutdown(context.Background())
	return
}
