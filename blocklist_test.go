package main

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleHosts = `# comment line
# Title: Steven Black sample

127.0.0.1 localhost
127.0.0.1 localhost.localdomain
255.255.255.255 broadcasthost
::1 localhost
::1 ip6-localhost
fe80::1%lo0 localhost
# blank line follows

0.0.0.0 googleadservices.com
0.0.0.0 doubleclick.net   # inline comment
0.0.0.0 Tracker.Example.com
`

func writeTempHosts(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "hosts")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadBlocklist(t *testing.T) {
	bl, err := LoadBlocklist(writeTempHosts(t, sampleHosts))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		want bool
	}{
		{"googleadservices.com", true},
		{"doubleclick.net", true},
		{"example.com", false},
		{"localhost", false},
		{"broadcasthost", false},
		{"ip6-localhost", false},
	}
	for _, tc := range cases {
		if got := bl.Contains(tc.name); got != tc.want {
			t.Errorf("Contains(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestLoadBlocklistMissingFile(t *testing.T) {
	if _, err := LoadBlocklist(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected error for missing file")
	}
}
