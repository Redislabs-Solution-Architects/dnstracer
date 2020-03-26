package rules

import (
	"fmt"

	"github.com/Redislabs-Solution-Architects/dnstracer/collection"
	"github.com/gookit/color"
)

// Results : Returns information about all fo the checks performed
type Results struct {
	ResultA      bool
	ResultNS     bool
	ResultGlue   bool
	ResultAccess bool
}

func debugPrint(check string, result bool) {
	if result {
		color.Green.Printf("\t%20s: OK\n", check)
	} else {
		color.Red.Printf("\t%20s: ERROR\n", check)
	}
}

func Check(collection collection.Collection, dbg bool) Results {

	results := Results{}

	// Start tests

	/* Check to make sure both public DNS server results match
	   Check that the LocalDNS and one of the remotes match
	   Check that there is more than 1 A record
	*/

	if collection.PublicMatchA && collection.LocalMatchA && len(collection.LocalA) > 0 {
		results.ResultA = true
	} else {
		results.ResultA = false
	}
	if dbg {
		fmt.Printf("--------------------------------\n")
		debugPrint("A Record Test", results.ResultA)
	}

	/* Check to make sure that the public DNS server NS records match
	   Check to make sure the one of the public and the private NS record servers match
	   Check to make sure there are at least 1 NS server
	*/

	if collection.PublicMatchNS && collection.LocalMatchNS && len(collection.LocalNS) > 0 {
		results.ResultNS = true
	} else {
		results.ResultNS = false
	}
	if dbg {
		debugPrint("NS Record Test", results.ResultNS)
	}

	/* Check to make sure the public DNS server Glue records match
	   Check to make sure the one of the public and the private Glue record servers match
	   Check to make sure there the Glue record length matches the ns record length
	*/

	if collection.PublicMatchGlue && collection.LocalMatchGlue && (len(collection.LocalNS) == len(collection.LocalGlue)) && len(collection.LocalNS) > 0 {
		results.ResultGlue = true
	} else {
		results.ResultGlue = false
	}
	if dbg {
		debugPrint("Glue Record Test", results.ResultGlue)
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

	if dbg {
		debugPrint("NS Access Test", results.ResultAccess)
	}

	if dbg {
		color.Cyan.Printf("--------------------------------\nResults Debug:\n%+v\n", results)
	}

	return (results)
}
