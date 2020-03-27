package main

import (
	"flag"
	"fmt"

	"github.com/Redislabs-Solution-Architects/dnstracer/collection"
	"github.com/Redislabs-Solution-Architects/dnstracer/rules"
)

func main() {
	cluster := flag.String("cluster-fqdn", "", "The name of the redis cluster eg: redis-10000.foo.example.com")
	dbg := flag.Bool("debug", false, "Show debug information")
	flag.Parse()

	coll := collection.Collect(*cluster)
	results := rules.Check(coll, *dbg)
	if *dbg {
		fmt.Printf("--------------------------------\n")
		fmt.Printf("%+v\n", coll)
		fmt.Printf("%+v\n", results)
	}
}
