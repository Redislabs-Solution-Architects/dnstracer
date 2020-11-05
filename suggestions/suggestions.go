package suggestions

import (
	"fmt"

	"github.com/Redislabs-Solution-Architects/dnstracer/collection"
	"github.com/Redislabs-Solution-Architects/dnstracer/rules"
)

// Suggest : Retun suggestions for possible fixes
func Suggest(collection *collection.Collection, results rules.Results, cluster *string) {
	fmt.Printf("Suggestions for %s\n--------------------------------\n", *cluster)
	if results.ResultNS != true {
		SuggestNS(collection, cluster)
	}
	if results.ResultGlue != true {
		SuggestGlue(collection, cluster)
	}
	if results.ResultAccess != true {
		SuggestAccess(collection, cluster)
	}
	if results.ResultSOAMatch != true {
		SuggestSOA(collection, cluster)
	}
	if results.ResultA != true {
		SuggestA(collection, cluster)
	}
}
