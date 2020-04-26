package suggestions

import (
	"fmt"
	"os"

	"github.com/Redislabs-Solution-Architects/dnstracer/collection"
)

// SuggestGlue : Retun suggestions for possible fixes
func SuggestGlue(collection *collection.Collection, cluster *string) {
	fmt.Printf("Glue Record Suggestions\n--------------------------------\n")
	fmt.Println(collection)
	os.Exit(1)
}
