package main

import "flag"

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
