package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type Server struct {
	cfg    Config
	client *dns.Client
	block  *Blocklist
	cache  *Cache
}

func NewServer(cfg Config, block *Blocklist) *Server {
	return &Server{cfg: cfg, client: &dns.Client{}, block: block, cache: NewCache()}
}

func (s *Server) handle(w dns.ResponseWriter, req *dns.Msg) {
	// RFC allows multiple questions, but no real resolver uses them.
	if len(req.Question) != 1 {
		m := new(dns.Msg)
		m.SetRcode(req, dns.RcodeFormatError)
		_ = w.WriteMsg(m)
		return
	}
	q := req.Question[0]
	name := strings.ToLower(strings.TrimSuffix(q.Name, "."))
	if cached, ok := s.cache.Get(q); ok {
		cached.Id = req.Id
		_ = w.WriteMsg(cached)
		return
	}
	if s.block != nil && s.block.Contains(name) {
		if s.cfg.Verbose {
			log.Printf("block %s", name)
		}
		m := new(dns.Msg)
		m.SetRcode(req, dns.RcodeNameError)
		_ = w.WriteMsg(m)
		return
	}
	resp, _, err := s.client.Exchange(req, s.cfg.Upstream)
	if err != nil {
		if s.cfg.Verbose {
			log.Printf("upstream error: %v", err)
		}
		m := new(dns.Msg)
		m.SetRcode(req, dns.RcodeServerFailure)
		_ = w.WriteMsg(m)
		return
	}
	if len(resp.Answer) > 0 {
		ttl := time.Duration(resp.Answer[0].Header().Ttl) * time.Second
		s.cache.Set(q, resp, ttl)
	}
	_ = w.WriteMsg(resp)
}

func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	mux := dns.NewServeMux()
	mux.HandleFunc(".", s.handle)
	srv := &dns.Server{Addr: addr, Net: "udp", Handler: mux}
	log.Printf("dns listening on %s udp, upstream=%s", addr, s.cfg.Upstream)
	return srv.ListenAndServe()
}
