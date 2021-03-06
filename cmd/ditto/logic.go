package main

import (
	"fmt"
	"github.com/domainr/whois"
	"github.com/evilsocket/islazy/async"
	"github.com/jpillora/go-tld"
	whoisparser "github.com/likexian/whois-parser-go"
	"golang.org/x/net/idna"
	"net"
)

func genEntries(parsed *tld.URL) {
	for i, c := range parsed.Domain {
		if substitutes, found := dictionary[c]; found {
			for _, sub := range substitutes {
				entries = append(entries, &Entry{
					Domain: fmt.Sprintf("%s%s%s.%s", parsed.Domain[:i], sub, parsed.Domain[i+1:], parsed.TLD),
				})
				if limit > 0 && len(entries) == limit {
					return
				}
			}
		}
	}
}

func isAvailable(domain string) (bool, *whoisparser.WhoisInfo) {
	req, err := whois.NewRequest(domain)
	if err != nil {
		return true, nil
	}

	resp, err := whois.DefaultClient.Fetch(req)
	if err != nil {
		return true, nil
	}

	parsed, err := whoisparser.Parse(string(resp.Body))
	if err != nil {
		return true, nil
	}

	return false, &parsed
}

func processEntry(arg async.Job) {
	defer progress.Increment()

	entry := arg.(*Entry)
	entry.Available, entry.Whois = isAvailable(entry.Domain)
	entry.Ascii, _ = idna.ToASCII(entry.Domain)
	// some whois might only be accepting ascii encoded domain names
	if entry.Available {
		entry.Available, entry.Whois = isAvailable(entry.Ascii)
	}

	if !entry.Available {
		entry.Addresses, _ = net.LookupHost(entry.Ascii)
		uniq := make(map[string]bool)
		for _, addr := range entry.Addresses {
			names, _ := net.LookupAddr(addr)
			for _, name := range names {
				uniq[name] = true
			}
		}
		for name, _ := range uniq {
			entry.Names = append(entry.Names, name)
		}
	}
}
