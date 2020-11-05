package rules

import (
	"fmt"
	"reflect"

	"github.com/Redislabs-Solution-Architects/dnstracer/collection"
	"github.com/gookit/color"
)

// Results : Returns information about all fo the checks performed
type Results struct {
	ResultA        bool
	ResultNS       bool
	ResultGlue     bool
	ResultAccess   bool
	ResultSOAMatch bool
}

func debugPrint(check string, result bool) {
	if result {
		color.Green.Printf("\t%20s: OK\n", check)
	} else {
		color.Red.Printf("\t%20s: ERROR\n", check)
	}
}

// Check : Given the collected data return a Results Struct
func Check(collection *collection.Collection, dbg, suggest bool) Results {

	results := Results{}

	// Start tests

	/* Check to make sure that the public DNS server NS records match
	   Check to make sure the one of the public and the private NS record servers match
	   Check to make sure there are at least 1 NS server
	*/

	collection.PublicMatchNS = reflect.DeepEqual(collection.DNS1NS, collection.DNS2NS)
	collection.LocalMatchNS = reflect.DeepEqual(collection.DNS1NS, collection.LocalNS)
	if collection.PublicMatchNS && collection.LocalMatchNS && len(collection.LocalNS) > 0 {
		results.ResultNS = true
	} else {
		results.ResultNS = false
	}

	/* Check to make sure the public DNS server Glue records match
	   Check to make sure the one of the public and the private Glue record servers match
	   Check to make sure there the Glue record length matches the ns record length
	*/

	collection.PublicMatchGlue = reflect.DeepEqual(collection.DNS1Glue, collection.DNS2Glue)
	collection.LocalMatchGlue = reflect.DeepEqual(collection.DNS1Glue, collection.LocalGlue)

	if collection.PublicMatchGlue && collection.LocalMatchGlue && (len(collection.LocalNS) == len(collection.LocalGlue)) && len(collection.LocalNS) > 0 {
		results.ResultGlue = true
	} else {
		results.ResultGlue = false
	}

	/* Check to make sure that we can access all of the name servers and the numbers match */

	results.ResultAccess = true
	for _, a := range collection.EndpointStatus {
		if a && results.ResultAccess {
		} else {
			results.ResultAccess = false
		}
	}
	if len(collection.EndpointStatus) != len(collection.LocalNS) || len(collection.EndpointStatus) < 1 {
		results.ResultAccess = false
	}

	/* Check to make sure both public DNS server results match
	   Check that the LocalDNS and one of the remotes match
	   Check that there is more than 1 A record
	*/

	collection.PublicMatchA = reflect.DeepEqual(collection.DNS1A, collection.DNS2A)
	collection.LocalMatchA = reflect.DeepEqual(collection.DNS1A, collection.LocalA)

	if collection.PublicMatchA && collection.LocalMatchA && len(collection.LocalA) > 0 && (len(collection.LocalA) == len(collection.DNS1A)) {
		results.ResultA = true
	} else {
		results.ResultA = false
	}

	// check to make sure the SOA records match the domain name we expect
	results.ResultSOAMatch = collection.SOAMatch

	// Show test results if suggest or debug
	if dbg || suggest {
		fmt.Printf("--------------------------------\n")
		debugPrint("NS Record Test", results.ResultNS)
		debugPrint("Glue Record Test", results.ResultGlue)
		debugPrint("NS Access Test", results.ResultAccess)
		debugPrint("SOA Match Test", results.ResultSOAMatch)
		debugPrint("A Record Test", results.ResultA)
		fmt.Printf("--------------------------------\n")
	}

	// only print datastructure if debug is on
	if dbg {
		color.Cyan.Printf("Results Debug:\n%+v\n", results)
	}

	return (results)
}
