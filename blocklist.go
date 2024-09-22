package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
)

var preambleIPs = map[string]struct{}{
	"::1": {}, "fe80::": {}, "ff00::": {}, "ff02::": {},
	"255.255.255.255": {},
}

var preambleNames = map[string]struct{}{
	"0.0.0.0": {}, "localhost": {}, "localhost.localdomain": {},
	"local": {}, "broadcasthost": {},
	"ip6-localhost": {}, "ip6-loopback": {}, "ip6-localnet": {},
	"ip6-mcastprefix": {}, "ip6-allnodes": {}, "ip6-allrouters": {}, "ip6-allhosts": {},
}

// Blocklist uses a Bloom filter as a fast negative check, then confirms a
// "maybe" hit with binary search over a sorted slice.
type Blocklist struct {
	domains []string
	bf      *bloom.BloomFilter
}

func LoadBlocklist(source string) (*Blocklist, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		path, err := downloadBlocklist(source)
		if err != nil {
			return nil, err
		}
		source = path
	}
	f, err := os.Open(source)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var domains []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<16), 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if _, skip := preambleIPs[fields[0]]; skip {
			continue
		}
		name := strings.ToLower(fields[1])
		if _, skip := preambleNames[name]; skip {
			continue
		}
		domains = append(domains, name)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	sort.Strings(domains)
	domains = dedupSorted(domains)
	n := uint(len(domains))
	if n == 0 {
		n = 1
	}
	bf := bloom.NewWithEstimates(n, 0.001)
	for _, d := range domains {
		bf.AddString(d)
	}
	return &Blocklist{domains: domains, bf: bf}, nil
}

func downloadBlocklist(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}
	tmp, err := os.CreateTemp("", "sinkhole-blocklist-*.txt")
	if err != nil {
		return "", err
	}
	defer tmp.Close()
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return "", err
	}
	return tmp.Name(), nil
}

func dedupSorted(in []string) []string {
	if len(in) < 2 {
		return in
	}
	out := in[:1]
	for _, s := range in[1:] {
		if s != out[len(out)-1] {
			out = append(out, s)
		}
	}
	return out
}

func (b *Blocklist) Contains(name string) bool {
	start := time.Now()
	defer func() { blocklistLatency.Observe(time.Since(start).Seconds()) }()
	name = strings.ToLower(name)
	if !b.bf.TestString(name) {
		return false
	}
	i := sort.SearchStrings(b.domains, name)
	return i < len(b.domains) && b.domains[i] == name
}

func (b *Blocklist) Len() int { return len(b.domains) }
