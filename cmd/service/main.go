package main

import (
	"crypto/tls"
	"flag"
	"github.com/cadence-workflow/starlark-worker/internal/cadence"
	"github.com/cadence-workflow/starlark-worker/internal/temporal"
	"github.com/cadence-workflow/starlark-worker/plugin"
	"github.com/cadence-workflow/starlark-worker/service"
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

	backend := cadence.GetBackend()
	service := &service.Service{
		Plugins:        plugin.Registry,
		ClientTaskList: opt.ClientTaskList,
		Backend:        backend,
	}
	if opt.Backend == "temporal" {
		backend = temporal.GetBackend()
	}

	serviceWorker := backend.RegisterWorker(opt.CadenceURL, opt.CadenceDomain, opt.CadenceTaskList, logger)
	service.Register(serviceWorker)

	if err := serviceWorker.Start(); err != nil {
		logger.Fatal("Start", zap.Error(err))
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	logger.Info("Server started. Press CTRL+C to exit.")

	<-sig
	serviceWorker.Stop()
	logger.Info("EXIT.")
}
