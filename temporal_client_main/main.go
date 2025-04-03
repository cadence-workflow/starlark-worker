package temporal_client_main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	temporalclient "github.com/cadence-workflow/starlark-worker/client/temporal_client"
	"github.com/cadence-workflow/starlark-worker/star"

	"go.starlark.net/starlark"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
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
	return "[" + strings.Join(*r, ", ") + "]"
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
	fs.StringVar(&rootDir, "root-dir", ".", "Root directory to build a package from.")
	fs.StringVar(&file, "file", "", "main file to create a package for")
	if err := fs.Parse(args); err != nil {
		log.Fatal(err)
	}
	if file == "" {
		log.Fatal("ERROR: --file required")
	}
	if err := temporalclient.Package(rootDir, file, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func __run__(_args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)

	var _package, file, function, args, kwargs, temporalEndpoint, namespace, taskqueue string
	var env StringSliceValue

	fs.StringVar(&_package, "package", ".", "Path to package file or directory")
	fs.StringVar(&file, "file", "", "Entrypoint .star file")
	fs.StringVar(&function, "function", "main", "Entrypoint function name")
	fs.StringVar(&args, "args", "[]", "Function positional args (JSON array)")
	fs.StringVar(&kwargs, "kwargs", "[]", "Function keyword args (JSON array of arrays)")
	fs.Var(&env, "env", "Environment variables (KEY=VALUE format)")
	fs.StringVar(&temporalEndpoint, "temporal-url", "localhost:7233", "")
	fs.StringVar(&namespace, "namespace", "default", "")
	fs.StringVar(&taskqueue, "taskqueue", "default", "")

	if err := fs.Parse(_args); err != nil {
		log.Fatal(err)
	}

	if file == "" {
		log.Fatal("ERROR: --file required")
	}

	var _tar []byte
	var err error
	if _package == "-" {
		_tar, err = io.ReadAll(os.Stdin)
	} else if fi, err := os.Stat(_package); err == nil && fi.IsDir() {
		buf := bytes.Buffer{}
		if err := temporalclient.Package(_package, file, &buf); err != nil {
			log.Fatal(err)
		}
		_tar = buf.Bytes()
		if file, err = filepath.Rel(_package, file); err != nil {
			log.Fatal(err)
		}
	} else {
		_tar, err = os.ReadFile(_package)
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Package size: %d bytes", len(_tar))

	var _env *starlark.Dict
	if len(env) > 0 {
		_env = &starlark.Dict{}
		for _, kv := range env {
			k, v, found := strings.Cut(kv, "=")
			if !found {
				log.Fatalf("Bad env format: %s", kv)
			}
			if err := _env.SetKey(starlark.String(k), starlark.String(v)); err != nil {
				log.Fatal(err)
			}
		}
	}

	var argsP starlark.Tuple
	if err = star.Decode([]byte(args), &argsP); err != nil {
		log.Fatal(err)
	}

	var kwargsP []starlark.Tuple
	if err = star.Decode([]byte(kwargs), &kwargsP); err != nil {
		log.Fatal(err)
	}

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	c, err := client.Dial(client.Options{
		HostPort:  temporalEndpoint,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	log.Printf("Running entrypoint: %s", function)
	if err := temporalclient.Run(_tar, file, function, argsP, kwargsP, _env, c, taskqueue); err != nil {
		log.Fatal(err)
	}
}
