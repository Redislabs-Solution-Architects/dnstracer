package suggestions

import (
	"fmt"
	"os"
	"strings"

	"github.com/Redislabs-Solution-Architects/dnstracer/collection"
)

// SuggestAccess : Retun suggestions for possible fixes
func SuggestAccess(collection *collection.Collection, cluster *string) {
	fmt.Printf("NS Access Suggestions\n--------------------------------\n")
	fmt.Println("Unable to query one or more of the delegated nameservers")
	fmt.Println("Please use the following commands to troubleshoot:")
	for _, i := range collection.LocalGlue {
		fmt.Printf("\tdig +noall +answer NS %s @%s\n", strings.Join(strings.Split(*cluster, ".")[1:], "."), i)
	}

	os.Exit(1)
}
