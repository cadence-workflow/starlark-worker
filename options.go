package main

import "flag"

type _Options struct {
	CadenceURL      string
	CadenceDomain   string
	CadenceTaskList string
	ClientTaskList  string
}

func (r *_Options) BindFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&r.CadenceURL,
		"cadence-url",
		"grpc://localhost:7833",
		"Cadence connection URL",
	)

	fs.StringVar(
		&r.CadenceDomain,
		"cadence-domain",
		"default",
		"Cadence domain",
	)
	fs.StringVar(
		&r.CadenceTaskList,
		"cadence-task-list",
		"default",
		"Cadence worker's TaskList",
	)
	fs.StringVar(
		&r.ClientTaskList,
		"client-task-list",
		"",
		"TaskList used by Cadence client to call user activities and workflows",
	)
}
