package main

import (
	"crypto/tls"
	"flag"
	"github.com/cadence-workflow/starlark-worker/cad"
	"github.com/cadence-workflow/starlark-worker/cadstar"
	"github.com/cadence-workflow/starlark-worker/plugin"
	"github.com/uber-go/tally"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opt := _Options{}
	opt.BindFlags(flags)
	if err := flags.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	z := zap.NewDevelopmentConfig()
	z.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	logger, err := z.Build()
	if err != nil {
		panic(err)
	}

	logger.Info("Options", zap.Any("Options", opt))

	cadInterface := cad.NewInterface(opt.CadenceURL, opt.CadenceService)

	cadWorker := worker.New(
		cadInterface,
		opt.CadenceDomain,
		opt.CadenceTaskList,
		worker.Options{
			MetricsScope: tally.NoopScope,
			Logger:       logger,
			DataConverter: &cadstar.DataConverter{
				Logger: logger,
			},
			ContextPropagators: []workflow.ContextPropagator{
				&cad.HeadersContextPropagator{},
			},
		},
	)

	service := &cadstar.Service{
		Plugins:        plugin.Registry,
		ClientTaskList: opt.ClientTaskList,
	}
	service.Register(cadWorker)

	if err := cadWorker.Start(); err != nil {
		logger.Fatal("Start", zap.Error(err))
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL)
	logger.Info("Server started. Press CTRL+C to exit.")

	select {
	case <-sig:
		cadWorker.Stop()
	}
	logger.Info("EXIT.")
}
