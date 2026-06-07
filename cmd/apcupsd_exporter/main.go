// Command apcupsd_exporter provides a Prometheus exporter for apcupsd.
package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mdlayher/apcupsd"
	apcupsdexporter "github.com/mdlayher/apcupsd_exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	telemetryAddr = flag.String("telemetry.addr", ":9162", "address for apcupsd exporter")
	metricsPath   = flag.String("telemetry.path", "/metrics", "URL path for surfacing collected metrics")

	apcupsdAddr    = flag.String("apcupsd.addr", ":3551", "address of apcupsd Network Information Server (NIS)")
	apcupsdNetwork = flag.String("apcupsd.network", "tcp", `network of apcupsd Network Information Server (NIS): typically "tcp", "tcp4", or "tcp6"`)
)

func main() {
	flag.Parse()

	if *apcupsdAddr == "" {
		slog.Error("address of apcupsd Network Information Server (NIS) must be specified", "flag", "-apcupsd.addr")
		os.Exit(1)
	}

	// Setup structured logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true}))

	fn := newClient(*apcupsdNetwork, *apcupsdAddr)

	prometheus.MustRegister(apcupsdexporter.NewWithLogger(fn, logger))

	mux := http.NewServeMux()
	mux.Handle(*metricsPath, promhttp.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, *metricsPath, http.StatusMovedPermanently)
	})
	mux.HandleFunc("/-/healthy", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/-/ready", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{Addr: *telemetryAddr, Handler: mux}

	logger.Info("starting apcupsd exporter", "addr", *telemetryAddr, "server", *apcupsdNetwork+"://"+*apcupsdAddr)

	// Start server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "err", err)
			os.Exit(1)
		}
	}()

	// Wait for termination signal and shutdown gracefully
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed", "err", err)
		os.Exit(1)
	}
}

func newClient(network, addr string) apcupsdexporter.ClientFunc {
	return func(ctx context.Context) (*apcupsd.Client, error) {
		return apcupsd.DialContext(ctx, network, addr)
	}
}
