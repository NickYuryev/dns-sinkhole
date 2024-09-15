package main

import "log"

func main() {
	cfg := parseFlags()
	block, err := LoadBlocklist(cfg.Blocklist)
	if err != nil {
		log.Fatalf("load blocklist: %v", err)
	}
	log.Printf("blocklist loaded: %d domains", block.Len())
	startMetricsServer(cfg.MetricsPort)
	s := NewServer(cfg, block)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
