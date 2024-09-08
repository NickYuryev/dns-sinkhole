package main

import (
	"bufio"
	"os"
	"strings"
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

type Blocklist struct {
	set map[string]struct{}
}

func NewBlocklist() *Blocklist {
	return &Blocklist{set: make(map[string]struct{})}
}

func LoadBlocklist(path string) (*Blocklist, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bl := NewBlocklist()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<16), 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// strip inline comment
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
		bl.set[name] = struct{}{}
	}
	return bl, sc.Err()
}

func (b *Blocklist) Contains(name string) bool {
	_, ok := b.set[strings.ToLower(name)]
	return ok
}

func (b *Blocklist) Len() int { return len(b.set) }
