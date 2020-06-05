package main

import (
	"github.com/pborman/getopt/v2"

	"fmt"
	"os"

	"github.com/Redislabs-Solution-Architects/dnstracer/collection"
	"github.com/Redislabs-Solution-Architects/dnstracer/rules"
	"github.com/Redislabs-Solution-Architects/dnstracer/suggestions"
)

// Name is the exported name of this application.
const Name = "dnstracer"

// Version is the current version of this application.
const Version = "0.0.3"

func main() {

	helpFlag := getopt.BoolLong("help", 'h', "display help")
	endpoint := getopt.StringLong("endpoint", 'e', "", "The name of the redis endpoint eg: redis-10000.foo.example.com")
	dbg := getopt.BoolLong("debug", 'd', "Enable debug output")
	suggest := getopt.BoolLong("suggest", 's', "Suggest possible fixes")
	getopt.Parse()

	if *helpFlag || *endpoint == "" {
		getopt.PrintUsage(os.Stderr)
		os.Exit(1)
	}

	coll := collection.Collect(*endpoint)
	results := rules.Check(coll, *dbg, *suggest)

	if *dbg {
		fmt.Printf("--------------------------------\n")
		fmt.Printf("%+v\n", coll)
		fmt.Printf("%+v\n", results)
	}

	if results.ResultA && results.ResultGlue && results.ResultNS && results.ResultAccess {
		fmt.Println("OK")
	} else if *suggest {
		suggestions.Suggest(coll, results, endpoint)
	} else {
		fmt.Println("Error - run with --debug for more information or run with --suggest for hints on how to fix")

	}
}
