package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

var commands = map[string]struct {
	nargs  int
	f      func(url string, args []string)
	argstr string
}{
	"upload":   {-1, uploadCommand, "/src/dir /dest/dir"},
	"download": {-1, downloadCommand, "/src/dir /dest/dir"},
	"ls":       {0, lsCommand, "[path]"},
	"rm":       {0, rmCommand, "path"},
	"info":     {0, infoCommand, ""},
}

func init() {
	log.SetFlags(log.Lmicroseconds)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"Usage:\n  %s http://cbfs:8484/ cmd [-opts] cmdargs\n",
			os.Args[0])

		fmt.Fprintf(os.Stderr, "\nCommands:\n")

		for k, v := range commands {
			fmt.Fprintf(os.Stderr, "  %s %s\n", k, v.argstr)
		}

		fmt.Fprintf(os.Stderr, "\n---- Subcommand Options ----\n")

		fmt.Fprintf(os.Stderr, "\nls:\n")
		lsFlags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nrm:\n")
		rmFlags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nupload:\n")
		uploadFlags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\ninfo:\n")
		infoFlags.PrintDefaults()
		os.Exit(1)
	}

}

func maybeFatal(err error, msg string, args ...interface{}) {
	if err != nil {
		log.Fatalf(msg, args...)
	}
}

func relativeUrl(u, path string) string {
	du, err := url.Parse(u)
	maybeFatal(err, "Error parsing url: %v", err)

	du.Path = path
	if du.Path[0] != '/' {
		du.Path = "/" + du.Path
	}

	return du.String()
}

func getJsonData(u string, into interface{}) error {
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("HTTP Error: %v", res.Status)
	}

	d := json.NewDecoder(res.Body)
	return d.Decode(into)
}

func verbose(v bool, f string, a ...interface{}) {
	if v {
		log.Printf(f, a...)
	}
}

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		flag.Usage()
	}

	u := flag.Arg(0)

	cmdName := flag.Arg(1)
	cmd, ok := commands[cmdName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown command: %v\n", cmdName)
		flag.Usage()
	}
	if cmd.nargs == 0 {
	} else if cmd.nargs < 0 {
		reqargs := -cmd.nargs
		if flag.NArg()-2 < reqargs {
			fmt.Fprintf(os.Stderr, "Incorrect arguments for %v\n", cmdName)
			flag.Usage()
		}
	} else {
		if flag.NArg()-2 != cmd.nargs {
			fmt.Fprintf(os.Stderr, "Incorrect arguments for %v\n", cmdName)
			flag.Usage()
		}
	}

	cmd.f(u, flag.Args()[2:])
}
