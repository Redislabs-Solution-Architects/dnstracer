package suggestions

import (
	"fmt"
	"strings"

	"github.com/Redislabs-Solution-Architects/dnstracer/collection"
)

// Suggest : Retun suggestions for possible fixes
func SuggestA(collection *collection.Collection, cluster *string) {
	fmt.Printf("A Record Suggestions\n--------------------------------\n")
	fmt.Println("The host", *cluster, "is not resolving properly")
	fmt.Printf("Check the user interface at https://%s:8443\n", strings.Join(strings.Split(*cluster, ".")[1:], "."))
	fmt.Println("After confirming the DB exists you can confirm it is working with the following command:")
	fmt.Printf("\tdig +noall +answer A %s @%s\n", *cluster, strings.Join(strings.Split(*cluster, ".")[1:], "."))

}
