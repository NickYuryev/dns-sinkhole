package main

import (
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// startTestServer runs the sinkhole on a free UDP port and returns its address.
func startTestServer(t *testing.T, bl *Blocklist, upstream string) string {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := pc.LocalAddr().(*net.UDPAddr).Port
	_ = pc.Close()

	cfg := Config{Port: port, Upstream: upstream}
	s := NewServer(cfg, bl)
	mux := dns.NewServeMux()
	mux.HandleFunc(".", s.handle)
	srv := &dns.Server{Addr: "127.0.0.1:" + strconv.Itoa(port), Net: "udp", Handler: mux}
	ready := make(chan struct{})
	srv.NotifyStartedFunc = func() { close(ready) }
	go func() { _ = srv.ListenAndServe() }()
	t.Cleanup(func() { _ = srv.Shutdown() })
	select {
	case <-ready:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not start")
	}
	return fmt.Sprintf("127.0.0.1:%d", port)
}

// startFakeUpstream answers any A query with 203.0.113.1 so tests do not hit the network.
func startFakeUpstream(t *testing.T) string {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := pc.LocalAddr().(*net.UDPAddr).Port
	_ = pc.Close()
	mux := dns.NewServeMux()
	mux.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(req)
		if len(req.Question) == 1 && req.Question[0].Qtype == dns.TypeA {
			rr, _ := dns.NewRR(req.Question[0].Name + " 60 IN A 203.0.113.1")
			m.Answer = append(m.Answer, rr)
		}
		_ = w.WriteMsg(m)
	})
	srv := &dns.Server{Addr: "127.0.0.1:" + strconv.Itoa(port), Net: "udp", Handler: mux}
	ready := make(chan struct{})
	srv.NotifyStartedFunc = func() { close(ready) }
	go func() { _ = srv.ListenAndServe() }()
	t.Cleanup(func() { _ = srv.Shutdown() })
	<-ready
	return fmt.Sprintf("127.0.0.1:%d", port)
}

func query(t *testing.T, addr, name string, qtype uint16) *dns.Msg {
	t.Helper()
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(name), qtype)
	c := &dns.Client{Timeout: 2 * time.Second}
	resp, _, err := c.Exchange(m, addr)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestServerBlockedReturnsNXDomain(t *testing.T) {
	bl, err := LoadBlocklist(writeTempHosts(t, sampleHosts))
	if err != nil {
		t.Fatal(err)
	}
	up := startFakeUpstream(t)
	addr := startTestServer(t, bl, up)
	resp := query(t, addr, "googleadservices.com", dns.TypeA)
	if resp.Rcode != dns.RcodeNameError {
		t.Fatalf("rcode=%d want NXDomain", resp.Rcode)
	}
}

func TestServerAllowedForwards(t *testing.T) {
	bl, err := LoadBlocklist(writeTempHosts(t, sampleHosts))
	if err != nil {
		t.Fatal(err)
	}
	up := startFakeUpstream(t)
	addr := startTestServer(t, bl, up)
	resp := query(t, addr, "example.com", dns.TypeA)
	if resp.Rcode != dns.RcodeSuccess {
		t.Fatalf("rcode=%d want success", resp.Rcode)
	}
	if len(resp.Answer) == 0 {
		t.Fatal("want at least one answer")
	}
}

func TestServerRejectsMultiQuestion(t *testing.T) {
	bl, err := LoadBlocklist(writeTempHosts(t, sampleHosts))
	if err != nil {
		t.Fatal(err)
	}
	addr := startTestServer(t, bl, "127.0.0.1:1") // unused
	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = true
	m.Question = []dns.Question{
		{Name: "example.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
		{Name: "example.org.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
	}
	c := &dns.Client{Timeout: 2 * time.Second}
	resp, _, err := c.Exchange(m, addr)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Rcode != dns.RcodeFormatError {
		t.Fatalf("rcode=%d want FormatError", resp.Rcode)
	}
}
