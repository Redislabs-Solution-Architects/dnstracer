package main

import (
	"github.com/pborman/getopt/v2"

	"fmt"
	"os"
	"strings"

	"github.com/Redislabs-Solution-Architects/dnstracer/collection"
	"github.com/Redislabs-Solution-Architects/dnstracer/rules"
	"github.com/Redislabs-Solution-Architects/dnstracer/suggestions"
)

// Name is the exported name of this application.
const Name = "dnstracer"

// Version is the current version of this application.
const Version = "0.0.6"

func main() {

	helpFlag := getopt.BoolLong("help", 'h', "display help")
	endpoint := getopt.StringLong("endpoint", 'e', "", "The name of the redis endpoint eg: redis-10000.foo.example.com")
	dbg := getopt.BoolLong("debug", 'd', "Enable debug output")
	intOnly := getopt.BoolLong("internal", 'i', "Use only internal resolvers (only on Linux)")
	suggest := getopt.BoolLong("suggest", 's', "Suggest possible fixes")
	getopt.Parse()

	if *helpFlag || *endpoint == "" {
		getopt.PrintUsage(os.Stderr)
		os.Exit(1)
	}

	if !strings.HasPrefix(strings.ToUpper(*endpoint), "REDIS") {
		fmt.Println("Endpoint name needs to start with redis.\nAn endpoint must already be created to test")
		os.Exit(1)
	}

	coll := collection.Collect(*endpoint, *intOnly)
	results := rules.Check(coll, *dbg, *suggest)

	if *dbg {
		fmt.Printf("--------------------------------\n")
		fmt.Printf("%+v\n", coll)
		fmt.Printf("%+v\n", results)
	}

	if results.ResultA && results.ResultGlue && results.ResultNS && results.ResultAccess && results.ResultSOAMatch {
		fmt.Println("OK")
	} else if *suggest {
		suggestions.Suggest(coll, results, endpoint)
	} else {
		fmt.Println("Error - run with --debug for more information or run with --suggest for hints on how to fix")

	}
}
