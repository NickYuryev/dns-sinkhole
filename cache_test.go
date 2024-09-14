package main

import (
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestCacheRoundTrip(t *testing.T) {
	c := NewCache()
	q := dns.Question{Name: "example.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET}
	m := new(dns.Msg)
	m.SetQuestion(q.Name, q.Qtype)
	rr, _ := dns.NewRR("example.com. 60 IN A 1.2.3.4")
	m.Answer = append(m.Answer, rr)
	c.Set(q, m, time.Minute)
	got, ok := c.Get(q)
	if !ok {
		t.Fatal("miss after set")
	}
	if len(got.Answer) != 1 {
		t.Fatalf("answers=%d want 1", len(got.Answer))
	}
}

func TestCacheMiss(t *testing.T) {
	c := NewCache()
	q := dns.Question{Name: "nope.example.", Qtype: dns.TypeA, Qclass: dns.ClassINET}
	if m, ok := c.Get(q); ok || m != nil {
		t.Fatalf("expected miss, got (%v, %v)", m, ok)
	}
}

func TestCacheKeySeparatesQtype(t *testing.T) {
	c := NewCache()
	qA := dns.Question{Name: "example.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET}
	qAAAA := dns.Question{Name: "example.com.", Qtype: dns.TypeAAAA, Qclass: dns.ClassINET}
	m := new(dns.Msg)
	m.SetQuestion(qA.Name, qA.Qtype)
	c.Set(qA, m, time.Minute)
	if _, ok := c.Get(qAAAA); ok {
		t.Fatal("AAAA should not hit A cache entry")
	}
}
