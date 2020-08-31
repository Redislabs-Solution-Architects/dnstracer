package collection

import (
	"context"
	"fmt"
	"net"
	"os"
	"reflect"
	"sort"
	"strings"
)

// Collection : a Struct returnin all the collected data from DNS servers
type Collection struct {
	LocalA          []string
	GoogleA         []string
	CFlareA         []string
	LocalNS         []string
	GoogleNS        []string
	CFlareNS        []string
	LocalGlue       []string
	GoogleGlue      []string
	CFlareGlue      []string
	PublicMatchA    bool
	LocalMatchA     bool
	PublicMatchNS   bool
	LocalMatchNS    bool
	PublicMatchGlue bool
	LocalMatchGlue  bool
	EndpointStatus  []bool
}

// Collect : grab all of the necessary information needed to make decisions
func Collect(cluster string, intOnly bool) *Collection {

	extResolvers := []string{"1.1.1.1:53", "8.8.8.8:53"}

	if intOnly {
		extResolvers = getDNSConf()
	}

	cfResolv := &net.Resolver{
		PreferGo: true,
		Dial:     dnsDial(extResolvers[0]),
	}
	gooResolv := &net.Resolver{
		PreferGo: true,
		Dial:     dnsDial(extResolvers[1]),
	}
	localResolv := &net.Resolver{
		PreferGo: true,
	}

	results := &Collection{}

	cfa, _ := cfResolv.LookupHost(context.Background(), cluster)
	cfg, _ := gooResolv.LookupHost(context.Background(), cluster)
	la, _ := localResolv.LookupHost(context.Background(), cluster)
	cfns, _ := cfResolv.LookupNS(context.Background(), strings.Join(strings.Split(cluster, ".")[1:], "."))
	goons, _ := gooResolv.LookupNS(context.Background(), strings.Join(strings.Split(cluster, ".")[1:], "."))
	localns, lerr := localResolv.LookupNS(context.Background(), strings.Join(strings.Split(cluster, ".")[1:], "."))
	if lerr != nil {
		// TODO: figure out a good way to find the resolver based on various OS's but probably not Windows
		x, y, _ := noDNSPropQuery(strings.Join(strings.Split(cluster, ".")[1:], "."), "127.0.0.1:53")
		for _, n := range x {
			results.LocalNS = append(results.LocalNS, n)
		}
		for _, n := range y {
			results.LocalGlue = append(results.LocalGlue, n)
		}
	} else {
		results.LocalNS = cleanNS(localns)
	}

	// Resolve all of the Glue records locally only if the slice is empty
	// If we couldn't resolve the localns above we do the no recurstion trick
	if len(results.LocalGlue) < 1 {
		for _, glu := range results.LocalNS {
			q, err := localResolv.LookupHost(context.Background(), glu)
			if err != nil {
				fmt.Println("ERR:", err)
			}
			q = cleanIPV6(q)
			for _, w := range q {
				results.LocalGlue = append(results.LocalGlue, w)
			}
		}
	}
	sort.Strings(results.LocalGlue)

	// Sort and clean all of the lookup results
	results.CFlareNS = cleanNS(cfns)
	results.GoogleNS = cleanNS(goons)
	results.CFlareA = cleanIPV6(cfa)
	results.GoogleA = cleanIPV6(cfg)
	results.LocalA = cleanIPV6(la)

	// Resolve all of the Glue records on Google
	for _, glu := range results.GoogleNS {
		q, err := gooResolv.LookupHost(context.Background(), glu)
		if err != nil {
			fmt.Println("ERR:", err)
		}
		q = cleanIPV6(q)
		for _, w := range q {
			results.GoogleGlue = append(results.GoogleGlue, w)
		}
	}
	sort.Strings(results.GoogleGlue)

	// Resolve all of the Glue records on Cloudflare
	for _, glu := range results.CFlareNS {
		q, err := cfResolv.LookupHost(context.Background(), glu)
		if err != nil {
			fmt.Println("ERR:", err)
		}
		q = cleanIPV6(q)
		for _, w := range q {
			results.CFlareGlue = append(results.CFlareGlue, w)
		}
	}
	sort.Strings(results.CFlareGlue)

	// Ensure we can dig against all the NS
	if (reflect.DeepEqual(results.CFlareGlue, results.GoogleGlue) && reflect.DeepEqual(results.CFlareGlue, results.LocalGlue)) || (len(results.GoogleNS) == 0 && len(results.LocalNS) != 0) {
		for _, r := range results.LocalGlue {
			w := &net.Resolver{
				PreferGo: true,
				Dial:     dnsDial(fmt.Sprintf("%s:53", r)),
			}
			_, err := w.LookupNS(context.Background(), strings.Join(strings.Split(cluster, ".")[1:], "."))
			if err == nil {
				results.EndpointStatus = append(results.EndpointStatus, true)
			} else {
				results.EndpointStatus = append(results.EndpointStatus, false)
			}
		}
	} else {
		fmt.Println("Unexpected Error occured - please re-run with the --debug")
		os.Exit(1)
	}

	return (results)
}
