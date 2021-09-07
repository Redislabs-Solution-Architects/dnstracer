package collection

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// Parse the DNS configuration to get possible server list
func getDNSConf() []string {
	servers := []string{}
	_, err := os.Stat("/etc/resolv.conf")
	if err != nil {
		log.Fatal(err)
	}
	j, _ := dns.ClientConfigFromFile("/etc/resolv.conf")

	servers = append(servers, fmt.Sprintf("%s:53", j.Servers[0]))
	if len(servers) < 2 {
		servers = append(servers, fmt.Sprintf("%s:53", j.Servers[0]))
	} else {
		servers = append(servers, fmt.Sprintf("%s:53", j.Servers[1]))
	}

	return servers

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

// DNS dialer function
func dnsDial(dnsServer string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{
			Timeout: time.Millisecond * time.Duration(2500),
		}
		return d.DialContext(ctx, "udp", dnsServer)
	}
}

// digSOA function

func matchSOA(dnsServer, domain string) bool {
	r := false
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(fmt.Sprintf("%s.", domain), dns.TypeSOA)
	soa, _, err := c.Exchange(m, dnsServer)
	if err != nil {
		log.Fatal("Unable to determine the SOA record for: ", domain, " from ", dnsServer)
	}
	if len(soa.Answer) > 0 {
		rr := strings.Split(soa.Answer[0].String(), "\t")
		ns := strings.Split(rr[4], " ")[0]
		if fmt.Sprintf("ns.%s.", domain) == ns {
			r = true
		}
	}
	return r

}
