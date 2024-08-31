package main

import (
	"fmt"
	"log"

	"github.com/miekg/dns"
)

type Server struct {
	cfg    Config
	client *dns.Client
}

func NewServer(cfg Config) *Server {
	return &Server{cfg: cfg, client: &dns.Client{}}
}

func (s *Server) handle(w dns.ResponseWriter, req *dns.Msg) {
	// RFC allows multiple questions, but no real resolver uses them.
	if len(req.Question) != 1 {
		m := new(dns.Msg)
		m.SetRcode(req, dns.RcodeFormatError)
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
	_ = w.WriteMsg(resp)
}

func (s *Server) ListenAndServe() error {
	dns.HandleFunc(".", s.handle)
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	srv := &dns.Server{Addr: addr, Net: "udp"}
	log.Printf("dns listening on %s udp, upstream=%s", addr, s.cfg.Upstream)
	return srv.ListenAndServe()
}
