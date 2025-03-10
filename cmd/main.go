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
	"time"

	"github.com/arraial/pipo-dispatcher/internal/settings"
	"github.com/arraial/pipo-dispatcher/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var version string = "latest"

func main() {
	var config = settings.InitConfig()

	log.Println(config.GetString("telemetry.service"))
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

	handleFunc("/livez", livezHandler)
	handleFunc("/readyz", readyzHandler)

	return otelhttp.NewHandler(mux, "/")
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
		Addr:         ":8080",
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
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
