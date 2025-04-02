package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/cadence-workflow/starlark-worker/cad"
	"github.com/cadence-workflow/starlark-worker/client"
	"github.com/cadence-workflow/starlark-worker/internal/temporal"
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/cadence-workflow/starlark-worker/star"
	"github.com/uber-go/tally"
	"go.starlark.net/starlark"
	tempclient "go.temporal.io/sdk/client"
	cadenceclient "go.uber.org/cadence/client"
	"go.uber.org/zap"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const help = `
Usage: %s [OPTIONS] TARGET [ARGS]...

Options:
  --help  Show this message and exit.

Targets:
  package    Create a starlark package file.
  run        Run a starlark package.

`

type StringSliceValue []string

func (r *StringSliceValue) String() string {
	var csv = ""
	for _, v := range *r {
		csv += v + ", "
	}
	return "[" + csv + "]"
}

func (r *StringSliceValue) Set(value string) error {
	*r = append(*r, value)
	return nil
}

var targets = map[string]func(args []string){
	"package": __package__,
	"run":     __run__,
}

func main() {
	log.SetFlags(0)
	if len(os.Args) == 0 {
		log.Fatal("ERROR: args undefined")
	}
	_help := fmt.Sprintf(help, os.Args[0])

	if len(os.Args) < 2 {
		log.Fatalf("%s ERROR: target undefined", _help)
	}
	arg1 := os.Args[1]
	if arg1 == "--help" || arg1 == "-h" {
		log.Printf(_help)
		return
	}
	target := targets[arg1]
	if target == nil {
		log.Fatalf("%s ERROR: unsupported target: %s", _help, arg1)
	}
	target(os.Args[2:])
}

func __package__(args []string) {

	fs := flag.NewFlagSet("tar", flag.ExitOnError)

	var rootDir, file string

	fs.StringVar(&rootDir, "root-dir", ".", "Root directory to build a package (tar archive) from.")
	fs.StringVar(&file, "file", "", "main file to create a package for")

	if err := fs.Parse(args); err != nil {
		log.Fatal(err)
	}

	log.Printf("rootDir: %s", rootDir)
	log.Printf("file: %s", file)

	if file == "" {
		log.Fatal("ERROR: --file required")
	}

	if err := cadenceclient.Package(rootDir, file, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func __run__(_args []string) {

	fs := flag.NewFlagSet("run", flag.ExitOnError)

	var _package, file, function, args, kwargs, endpoint, namespace, taskQueue string
	var env StringSliceValue

	fs.StringVar(&_package, "package", ".", "Package path. Package is a *.tar.gz file used to aggregate *.star files and associated resources into one file for distribution. There are 3 ways to specify the package: 1) file path (reads package directly from the file); 2) directory path (builds package from the directory); 3) - (single hyphen, reads package from stdin)")
	fs.StringVar(&file, "file", "", "Entry point *.star file that defines a function to run (see --function). This file must exist in the package file (see --package-root, --package).")
	fs.StringVar(&function, "function", "main", "Entry point function to run. The function must be defined in the entry point *.star file (see --file).")
	fs.StringVar(&args, "args", "[]", "Function's positional arguments. Format: JSON array. Example: '[\"foo\", 100, true]'")
	fs.StringVar(&kwargs, "kwargs", "[]", "Function's keyword arguments.Format: JSON array or arrays. Example: [[\"country_code\", \"US\"], [\"item_id\", 101]]")
	fs.Var(&env, "env", "Environment variables to be set for the run")
	fs.StringVar(&endpoint, "cadence-url", "grpc://localhost:7833", "")
	fs.StringVar(&namespace, "namespace", "default", "")
	fs.StringVar(&taskQueue, "taskQueue", "default", "")

	if err := fs.Parse(_args); err != nil {
		log.Fatal(err)
	}

	_help := fmt.Sprintf("Run `%s run --help` for help.", os.Args[0])

	if file == "" {
		log.Fatalf("ERROR: Required: --file. %s", _help)
	}

	var _tar []byte
	var err error
	if _package == "-" {
		log.Printf("Read package from: stdin")
		if _tar, err = io.ReadAll(os.Stdin); err != nil {
			log.Fatal(err)
		}
	} else {

		var fi os.FileInfo

		if f, err := os.Open(_package); err != nil {
			log.Fatal(err)
		} else if fi, err = f.Stat(); err != nil {
			log.Fatal(err)
		}
		if fi.IsDir() {

			log.Printf("Create package: %s, file: %s", _package, file)
			buf := bytes.Buffer{}
			if err := cadenceclient.Package(_package, file, &buf); err != nil {
				log.Fatal(err)
			}
			_tar = buf.Bytes()
			if file, err = filepath.Rel(_package, file); err != nil {
				log.Fatal(err)
			}
			log.Printf("Updated file: %s", file)

		} else {
			log.Printf("Read package: %s", _package)
			if _tar, err = os.ReadFile(_package); err != nil {
				log.Fatal(err)
			}
		}
	}
	log.Printf("Package size: %d bytes", len(_tar))

	log.Printf("Parse --env: %s", env)
	var _env *starlark.Dict
	if len(env) > 0 {
		_env = &starlark.Dict{}
		sep := "="
		for _, kv := range env {
			k, v, found := strings.Cut(kv, sep)
			if !found {
				log.Fatalf("ERROR: Bad env format: separator '%s' not found: %s", sep, kv)
			}
			if err := _env.SetKey(starlark.String(k), starlark.String(v)); err != nil {
				log.Fatal(err)
			}
		}
	}
	log.Printf("Env: %s", _env)

	log.Printf("Parse --args: %s", args)
	var argsP starlark.Tuple
	if err = star.Decode([]byte(args), &argsP); err != nil {
		log.Fatal(err)
	}
	for i, arg := range argsP {
		log.Printf("arg[%d]: %T: %s", i, arg, arg.String())
	}

	log.Printf("Parse --kwargs: %s", kwargs)
	var kwargsP []starlark.Tuple
	if err = star.Decode([]byte(kwargs), &kwargsP); err != nil {
		log.Fatal(err)
	}
	for _, arg := range kwargsP {
		log.Printf("keyword[%s]: %T: %s", arg[0].String(), arg[1], arg[1].String())
	}

	log.Printf("Cadence endpoint: %s, domain: %s, tasklist: %s", cadenceEndpoint, domain, tasklist)

	z := zap.NewDevelopmentConfig()
	z.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logger, err := z.Build()
	if err != nil {
		log.Fatal(err)
	}

	temporalInterface, err := temporal.NewClient(endpoint, namespace)
	if err != nil {
		log.Fatal(err)
	}
	cadenceCli := temporal.NewClient(temporalInterface, namespace, &cadenceclient.Options{
		MetricsScope: tally.NoopScope,
		DataConverter: &temporal.DataConverter{
			Logger: logger,
		},
	})

	log.Printf("Entrypoint function: %s", function)
	if err := cadenceclient.Run(_tar, file, function, argsP, kwargsP, _env, cadenceCli, taskQueue); err != nil {
		log.Fatal(err)
	}
}
