package suggestions

import (
	"fmt"
	"os"
	"strings"

	"github.com/Redislabs-Solution-Architects/dnstracer/collection"
)

// SuggestSOA : Retun suggestions for possible fixes
func SuggestSOA(collection *collection.Collection, cluster *string) {
	fmt.Printf("SOA Record Suggestions\n--------------------------------\n")

	if collection.SOAMatch == false {
		fmt.Println("The name servers did not return the proper SOA record")
		fmt.Println("Make sure the cluster name matches the DNS domain")
		fmt.Printf("\tCluster FQDN Should be %s\n", strings.Join(strings.Split(*cluster, ".")[1:], "."))
		os.Exit(1)
	}
}
