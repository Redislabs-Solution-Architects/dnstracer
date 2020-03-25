package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"reflect"
	"sort"
	"strings"
	"time"
)

type checks struct {
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
	PublicMatchGlue bool
	LocalMatchGlue  bool
	EndpointStatus  []bool
}

func dnsDial(dnsServer string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{
			Timeout: time.Millisecond * time.Duration(10000),
		}
		return d.DialContext(ctx, "udp", dnsServer)
	}
}

func cleanNS(l []*net.NS) []string {
	var r []string
	for _, i := range l {
		r = append(r, i.Host)
	}
	sort.Strings(r)
	return (r)

}

func main() {
	cluster := flag.String("cluster-fqdn", "", "The name of the redis cluster eg: redis-10000.foo.example.com")
	flag.Parse()
	cfResolv := &net.Resolver{
		PreferGo: true,
		Dial:     dnsDial("1.1.1.1:53"),
	}
	gooResolv := &net.Resolver{
		PreferGo: true,
		Dial:     dnsDial("8.8.8.8:53"),
	}
	localResolv := &net.Resolver{
		PreferGo: true,
	}

	results := checks{}

	results.CFlareA, _ = cfResolv.LookupHost(context.Background(), *cluster)
	results.GoogleA, _ = gooResolv.LookupHost(context.Background(), *cluster)
	results.LocalA, _ = localResolv.LookupHost(context.Background(), *cluster)
	cfns, _ := cfResolv.LookupNS(context.Background(), strings.Join(strings.Split(*cluster, ".")[1:], "."))
	goons, _ := gooResolv.LookupNS(context.Background(), strings.Join(strings.Split(*cluster, ".")[1:], "."))
	localns, _ := localResolv.LookupNS(context.Background(), strings.Join(strings.Split(*cluster, ".")[1:], "."))

	results.CFlareNS = cleanNS(cfns)
	results.GoogleNS = cleanNS(goons)
	results.LocalNS = cleanNS(localns)
	sort.Strings(results.LocalA)
	sort.Strings(results.CFlareA)
	sort.Strings(results.GoogleA)
	results.PublicMatchA = reflect.DeepEqual(results.CFlareA, results.GoogleA)
	results.LocalMatchA = reflect.DeepEqual(results.CFlareA, results.LocalA)

	// Resolve all of the Glue records locally
	for _, glu := range results.LocalNS {
		q, err := localResolv.LookupHost(context.Background(), glu)
		if err != nil {
			fmt.Println("ERR:", err)
		}
		for _, w := range q {
			results.LocalGlue = append(results.LocalGlue, w)
		}
	}
	sort.Strings(results.LocalGlue)

	// Resolve all of the Glue records on Google
	for _, glu := range results.GoogleNS {
		q, err := gooResolv.LookupHost(context.Background(), glu)
		if err != nil {
			fmt.Println("ERR:", err)
		}
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
		for _, w := range q {
			results.CFlareGlue = append(results.CFlareGlue, w)
		}
	}
	sort.Strings(results.CFlareGlue)

	results.PublicMatchGlue = reflect.DeepEqual(results.CFlareGlue, results.GoogleGlue)
	results.LocalMatchGlue = reflect.DeepEqual(results.CFlareGlue, results.LocalGlue)

	fmt.Printf("%+v\n", results)
}
