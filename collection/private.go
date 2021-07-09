package collection

import (
	"fmt"
	"net"
)

func ScanValidIPs(a ...[]string) bool {
	res := true
	for _, g := range a {
		for _, ip := range g {
			if !isValidIP(ip) {
				res = false
			}
		}
	}
	fmt.Println(res)
	return res
}

func isValidIP(ipaddr string) bool {
	var invalidIPBlocks []*net.IPNet

	ip := net.ParseIP(ipaddr)

	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		invalidIPBlocks = append(invalidIPBlocks, block)
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}

	for _, block := range invalidIPBlocks {
		if block.Contains(ip) {
			return false
		}
	}
	return true

}
