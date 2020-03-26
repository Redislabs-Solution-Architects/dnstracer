package main

import (
    "context"
    "flag"
    "fmt"
    "github.com/gookit/color"
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
    PublicMatchNS   bool
    LocalMatchNS    bool
    PublicMatchGlue bool
    LocalMatchGlue  bool
    EndpointStatus  []bool
    ResultA         bool
    ResultNS        bool
    ResultGlue      bool
    ResultAccess    bool
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

func debugPrint(check string, result bool) {
    if result {
        color.Green.Printf("\t%20s: OK\n", check)
    } else {
        color.Red.Printf("\t%20s: ERROR\n", check)
    }
}

func main() {
    cluster := flag.String("cluster-fqdn", "", "The name of the redis cluster eg: redis-10000.foo.example.com")
    dbg := flag.Bool("debug", false, "Show debug information")
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
    results.PublicMatchNS = reflect.DeepEqual(results.CFlareNS, results.GoogleNS)
    results.LocalMatchNS = reflect.DeepEqual(results.CFlareNS, results.LocalNS)

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

    if results.PublicMatchGlue && results.LocalMatchGlue {
        for _, r := range results.LocalGlue {
            w := &net.Resolver{
                PreferGo: true,
                Dial:     dnsDial(fmt.Sprintf("%s:53", r)),
            }
            _, err := w.LookupNS(context.Background(), strings.Join(strings.Split(*cluster, ".")[1:], "."))
            if err == nil {
                results.EndpointStatus = append(results.EndpointStatus, true)
            } else {
                results.EndpointStatus = append(results.EndpointStatus, false)
            }
        }
    }

    // Start tests

    /* Check to make sure both public DNS server results match
       Check that the LocalDNS and one of the remotes match
       Check that there is more than 1 A record
    */

    if results.PublicMatchA && results.LocalMatchA && len(results.LocalA) > 0 {
        results.ResultA = true
    } else {
        results.ResultA = false
    }
    if *dbg { 
        fmt.Printf("--------------------------------\n") 
        debugPrint("A Record Test", results.ResultA)
    }

    /* Check to make sure that the public DNS server NS records match
       Check to make sure the one of the public and the private NS record servers match
       Check to make sure there are at least 1 NS server
    */

    if results.PublicMatchNS && results.LocalMatchNS && len(results.LocalNS) > 0 {
        results.ResultNS = true
    } else {
        results.ResultNS = false
    }
    if *dbg { debugPrint("NS Record Test", results.ResultNS)}

    /* Check to make sure the public DNS server Glue records match
       Check to make sure the one of the public and the private Glue record servers match
       Check to make sure there the Glue record length matches the ns record length
    */

    if results.PublicMatchGlue && results.LocalMatchGlue && (len(results.LocalNS) == len(results.LocalGlue)) && len(results.LocalNS) >0 {
        results.ResultGlue = true
    } else {
        results.ResultGlue = false
    }
    if *dbg { debugPrint("Glue Record Test", results.ResultGlue)}
    
    /* Check to make sure that we can access all of the name servers and the numbers match */

    results.ResultAccess = true
    for _, a := range results.EndpointStatus {
        if a && results.ResultAccess {} else {
            results.ResultAccess = false
        }
    }
    if ( len(results.EndpointStatus) != len(results.LocalNS) || len(results.EndpointStatus) < 1 ) {
            results.ResultAccess = false
    }

    if *dbg { debugPrint("NS Access Test", results.ResultAccess)}
    

    if *dbg {
        color.Cyan.Printf("--------------------------------\nResults Debug:\n%+v\n", results)
    }
}
