package collection

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/miekg/dns"
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

// Stolen from https://gist.github.com/sajal/23798b930edd51cb925ef15c6b237f13
func noDNSPropQuery(fqdn, nameserver string) ([]string, []string, error) {
	glue := []string{}
	ns := []string{}
	if fqdn[len(fqdn)-1] != '.' {
		fqdn = fqdn + "."
	}
	m := new(dns.Msg)
	m.SetQuestion(fqdn, 2)
	m.SetEdns0(4096, false)
	m.RecursionDesired = false
	udp := &dns.Client{Net: "udp", Timeout: time.Millisecond * time.Duration(2500)}
	in, _, err := udp.Exchange(m, nameserver)
	if err != nil {
		fmt.Println("Error:", fqdn, " ", err)
	} else {

		re := regexp.MustCompile(`^(?P<fqdn>\S+)\s+\d+\s+IN\s+A\s+(?P<ip>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
		for _, i := range in.Extra {
			res := re.FindStringSubmatch(i.String())
			if len(res) > 0 {
				glue = append(glue, res[2])
				ns = append(ns, res[1])
			}
		}
	}
	return ns, glue, err
}

func dnsDial(dnsServer string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{
			Timeout: time.Millisecond * time.Duration(2500),
		}
		return d.DialContext(ctx, "udp", dnsServer)
	}
}

// strip out any ipv6 IP addresses
func cleanIPV6(l []string) []string {
	var r []string
	for _, i := range l {
		matched, _ := regexp.MatchString(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`, i)
		if matched {
			r = append(r, i)
		}
	}
	sort.Strings(r)
	return (r)
}

// Sorts and removes all ipv6 address
func cleanNS(l []*net.NS) []string {
	var r []string
	for _, i := range l {
		r = append(r, i.Host)
	}
	sort.Strings(r)
	return (r)

}

// Collect : grab all of the necessary information needed to make decisions
func Collect(cluster string) *Collection {

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
		fmt.Println("WTF?")
	}

	return (results)
}
