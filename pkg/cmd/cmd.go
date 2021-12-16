package cmd

import (
	"context"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	_ "go.uber.org/automaxprocs"

	"github.com/wish/mongoproxy/pkg/mongoproxy"
	"github.com/wish/mongoproxy/pkg/mongoproxy/config"
)

var opts struct {
	Config      string        `long:"config" description:"path to config file" required:"true"`
	LogLevel    string        `long:"log-level" description:"Log level" default:"info"`
	MetricsBind string        `long:"metrics-bind" description:"address to bind metrics interface to" required:"true"`
	TermSleep   time.Duration `long:"term-sleep" description:"how long to wait on shutdown after getting a termination signal" default:"5s"`
	SentryDSN   string        `long:"sentry-dsn" env:"SENTRY_DSN"`
}

func Main() {
	// Wait for reload or termination signals. Start the handler for SIGHUP as
	// early as possible, but ignore it until we are ready to handle reloading
	// our config.
	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		// If the error was from the parser, then we can simply return
		// as Parse() prints the error already
		if _, ok := err.(*flags.Error); ok {
			os.Exit(1)
		}
		logrus.Fatalf("error parsing flags: %v", err)
	}

	// Use log level
	level, err := logrus.ParseLevel(opts.LogLevel)
	if err != nil {
		logrus.Fatalf("Unknown log level %s: %v", opts.LogLevel, err)
	}
	logrus.SetLevel(level)

	// Set the log format to have a reasonable timestamp
	formatter := &logrus.TextFormatter{
		FullTimestamp: true,
	}
	logrus.SetFormatter(formatter)

	if opts.SentryDSN != "" {
		sentry.Init(sentry.ClientOptions{
			Dsn: opts.SentryDSN,
		})
	}

	// Get config
	cfg, err := config.ConfigFromFile(opts.Config)
	if err != nil {
		logrus.Fatal(err)
	}

	// Start up the metrics server
	ml, err := net.Listen("tcp", opts.MetricsBind)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("Metrics bind started: %v", ml.Addr())
	mux := http.NewServeMux()

	ready := false
	go func() {
		mux.Handle("/metrics", promhttp.Handler())

		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		// TODO: better HC
		// This is a dumb liveliness check endpoint. Currently this checks
		// nothing and will always return 200 if the process is live.
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			if !ready {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		})
		http.Serve(ml, mux)
	}()

	// Start up the server
	l, err := net.Listen("tcp", cfg.BindAddr) // TODO: allow multiple?
	if err != nil {
		logrus.Fatal(err)
	}

	proxy, err := mongoproxy.NewProxy(l, cfg)
	if err != nil {
		logrus.Fatal(err)
	}

	go func() {
		ready = true
		if err := proxy.Serve(); err != nil && err != mongoproxy.ErrServerClosed {
			logrus.Fatal(err)
		}
	}()

	logrus.Infof("started, ready!")

	// wait for signals etc.
	for sig := range sigs {
		switch sig {
		case syscall.SIGHUP:
			logrus.Infof("TODO: Reloading config")
		case syscall.SIGTERM, syscall.SIGINT:
			ready = false
			logrus.Infof("received exit signal, starting graceful shutdown after %v", opts.TermSleep)
			time.Sleep(opts.TermSleep)
			logrus.Info("starting graceful shutdown NOW")
			if err := proxy.Shutdown(context.TODO()); err != nil {
				logrus.Errorf("Error shutting down: %v", err)
			}
			return
		default:
			logrus.Errorf("Uncaught signal: %v", sig)
		}
	}
}
