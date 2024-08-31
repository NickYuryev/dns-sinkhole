package main

import (
	"bufio"
	"os"
	"strings"
)

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
		bl.set[fields[1]] = struct{}{}
	}
	return bl, sc.Err()
}

func (b *Blocklist) Contains(name string) bool {
	_, ok := b.set[name]
	return ok
}

func (b *Blocklist) Len() int { return len(b.set) }
