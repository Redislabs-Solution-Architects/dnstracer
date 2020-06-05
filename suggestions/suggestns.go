package suggestions

import (
	"fmt"
	"os"
	"strings"

	"github.com/Redislabs-Solution-Architects/dnstracer/collection"
)

// SuggestNS : Retun suggestions for possible fixes
func SuggestNS(collection *collection.Collection, cluster *string) {
	fmt.Printf("NS Record Suggestions\n--------------------------------\n")

	if collection.PublicMatchNS && len(collection.LocalGlue) > 0 && collection.EndpointStatus[0] == false {
		fmt.Println("The name servers may resolve to private IPs or we are not able to connect to them")
		fmt.Println("Please try adjusting firewall rules or ensure the DNS servers are running")
		fmt.Println("To confirm they are answering run the following commands:")
		for _, i := range collection.LocalGlue {
			fmt.Printf("\tdig +noall +answer NS %s @%s\n", strings.Join(strings.Split(*cluster, ".")[1:], "."), i)
		}
		os.Exit(1)
	}

	// Delegation completely failed
	if collection.LocalMatchNS && collection.PublicMatchNS && len(collection.LocalNS) == 0 {
		fmt.Printf("Unable to find a name server for subdomain %s\n", strings.Join(strings.Split(*cluster, ".")[1:], "."))
		fmt.Printf("Please make sure that domain %s is delegating to %s\n",
			strings.Join(strings.Split(*cluster, ".")[2:], "."),
			strings.Join(strings.Split(*cluster, ".")[1:], "."))
		fmt.Println("Use the following commands to troubleshoot:")
		fmt.Printf("\tdig +noall +answer NS %s\n", strings.Join(strings.Split(*cluster, ".")[1:], "."))
		fmt.Printf("\tdig +noall +answer NS %s\n", strings.Join(strings.Split(*cluster, ".")[2:], "."))
		os.Exit(1)
	}

	// Inconsistent NS records
	if len(collection.LocalNS) > 0 && (collection.LocalMatchNS != true || collection.PublicMatchNS != true) {
		fmt.Printf("We get inconsistent answers for NS for domain %s\n", strings.Join(strings.Split(*cluster, ".")[1:], "."))
		fmt.Printf("Local: %s\n", strings.Join(collection.LocalNS, ", "))
		fmt.Printf("Google: %s\n", strings.Join(collection.GoogleNS, ", "))
		fmt.Printf("Cloudflare: %s\n", strings.Join(collection.CFlareNS, ", "))
		fmt.Println("see https://github.com/Redislabs-Solution-Architects/dnstracer/troubleshooting/nsmismatch.md")
		os.Exit(1)

	}

	// Local lookup fails
	if len(collection.LocalNS) == 0 {
		fmt.Printf("Locally unable to find domain %s\n", strings.Join(strings.Split(*cluster, ".")[1:], "."))
		fmt.Println("Please contact your local Administrator and have them troubleshoot with the following command on the same server:")
		fmt.Printf("\tdig +noall +answer NS %s\n", strings.Join(strings.Split(*cluster, ".")[1:], "."))
	}
}
