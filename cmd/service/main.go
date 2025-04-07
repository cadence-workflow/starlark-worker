package main

import (
	"crypto/tls"
	"flag"
	"github.com/cadence-workflow/starlark-worker/plugin"
	"github.com/cadence-workflow/starlark-worker/service"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type _Options struct {
	Backend         string
	CadenceURL      string
	CadenceDomain   string
	CadenceTaskList string
	ClientTaskList  string
}

func (r *_Options) BindFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&r.Backend,
		"backend",
		"cadence",
		"Workflow backend to use: 'cadence' or 'temporal'",
	)
	fs.StringVar(
		&r.CadenceURL,
		"url",
		"grpc://localhost:7833",
		"Cadence or Temporal connection URL",
	)

	fs.StringVar(
		&r.CadenceDomain,
		"domain",
		"default",
		"Cadence domain or Temporal namespace",
	)
	fs.StringVar(
		&r.CadenceTaskList,
		"task-list",
		"default",
		"Cadence worker's TaskList or Temporal Taskqueue",
	)
	fs.StringVar(
		&r.ClientTaskList,
		"client-task-list",
		"",
		"TaskList used by Cadence client to call user activities and workflows",
	)
}

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

	backend := service.CadenceBackend
	if opt.Backend == "temporal" {
		backend = service.TemporalBackend
	}
	workerService, err := service.NewService(plugin.Registry, opt.ClientTaskList, backend)
	if err != nil {
		panic(err)
	}
	serviceWorker := workerService.RegisterWorker(opt.CadenceURL, opt.CadenceDomain, opt.CadenceTaskList, logger)
	workerService.Register(serviceWorker)

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
