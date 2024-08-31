package main

import "flag"

type Config struct {
	Port        int
	MetricsPort int
	Upstream    string
	Blocklist   string
	Verbose     bool
}

func parseFlags() Config {
	var c Config
	flag.IntVar(&c.Port, "port", 8053, "UDP port for DNS")
	flag.IntVar(&c.MetricsPort, "metrics-port", 9090, "TCP port for Prometheus metrics")
	flag.StringVar(&c.Upstream, "upstream", "1.1.1.1:53", "upstream DNS resolver")
	flag.StringVar(&c.Blocklist, "blocklist", "./blocklist.txt", "path or http(s) URL to a hosts-format blocklist")
	flag.BoolVar(&c.Verbose, "verbose", false, "log all queries")
	flag.Parse()
	return c
}
