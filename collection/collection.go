package collection

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"sort"
	"strings"
)

// Collection : a Struct returnin all the collected data from DNS servers
type Collection struct {
	LocalA          []string
	DNS2A           []string
	DNS1A           []string
	LocalNS         []string
	DNS2NS          []string
	DNS1NS          []string
	LocalGlue       []string
	DNS2Glue        []string
	DNS1Glue        []string
	SOAMatch        bool
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

	dns1Resolv := &net.Resolver{
		PreferGo: true,
		Dial:     dnsDial(extResolvers[0]),
	}
	dns2Resolv := &net.Resolver{
		PreferGo: true,
		Dial:     dnsDial(extResolvers[1]),
	}
	localResolv := &net.Resolver{
		PreferGo: true,
	}

	results := &Collection{}

	// Check upstream NS server to make sure they are good - die if they contain private IPs
	usns, _ := dns1Resolv.LookupNS(context.Background(), strings.Join(strings.Split(cluster, ".")[2:], "."))

	_, y, _ := noDNSPropQuery(strings.Join(strings.Split(cluster, ".")[1:], "."), fmt.Sprintf("%s:53", cleanNS(usns)[0]))
	if !ScanValidIPs(y) {
		log.Fatal("The glue records: ", y, "contain a private or invalid IP address")
	}

	dns1a, _ := dns1Resolv.LookupHost(context.Background(), cluster)
	dns2a, _ := dns2Resolv.LookupHost(context.Background(), cluster)
	la, _ := localResolv.LookupHost(context.Background(), cluster)
	dns1ns, _ := dns1Resolv.LookupNS(context.Background(), strings.Join(strings.Split(cluster, ".")[1:], "."))
	dns2ns, _ := dns2Resolv.LookupNS(context.Background(), strings.Join(strings.Split(cluster, ".")[1:], "."))
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
	results.DNS1NS = cleanNS(dns1ns)
	results.DNS2NS = cleanNS(dns2ns)
	results.DNS1A = cleanIPV6(dns1a)
	results.DNS2A = cleanIPV6(dns2a)
	results.LocalA = cleanIPV6(la)

	// Resolve all of the Glue records on DNS2
	for _, glu := range results.DNS2NS {
		q, err := dns2Resolv.LookupHost(context.Background(), glu)
		if err != nil {
			fmt.Println("ERR:", err)
		}
		q = cleanIPV6(q)
		for _, w := range q {
			results.DNS2Glue = append(results.DNS2Glue, w)
		}
	}
	sort.Strings(results.DNS2Glue)

	// Resolve all of the Glue records on Cloudflare
	for _, glu := range results.DNS1NS {
		q, err := dns1Resolv.LookupHost(context.Background(), glu)
		if err != nil {
			fmt.Println("ERR:", err)
		}
		q = cleanIPV6(q)
		for _, w := range q {
			results.DNS1Glue = append(results.DNS1Glue, w)
		}
	}
	sort.Strings(results.DNS1Glue)

	// Ensure we can dig against all the NS
	if (reflect.DeepEqual(results.DNS1Glue, results.DNS2Glue) && reflect.DeepEqual(results.DNS1Glue, results.LocalGlue)) || (len(results.DNS2NS) == 0 && len(results.LocalNS) != 0) {
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

	// Check to see if the SOA record proves the DNS name matches the cluster name
	if len(results.DNS1A) > 0 {
		results.SOAMatch = matchSOA(
			fmt.Sprintf("%s:53", results.LocalNS[0]),
			strings.Join(strings.Split(cluster, ".")[1:], "."))
	}

	return (results)
}
